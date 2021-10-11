package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	v3rpc "go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

// watchStreamRequest is a union of the supported watch request operation types
type watchStreamRequest interface {
	toPB() *pb.WatchRequest
}

// watcherStream represents a registered watcher
type watcherStream struct {
	// initReq is the request that initiated this request
	initReq watchRequest

	// id is the registered watch id on the grpc stream
	id int64

	// buf holds all events received from etcd but not yet consumed by the client
	buf []*WatchResponse
	// outc publishes watch responses to subscriber
	outc chan WatchResponse

	// receiveChan buffers watch responses before publishing
	receiveChan chan *WatchResponse

	// closing is set to true when stream should be scheduled to shutdown.
	closing bool

	// donec closes when the watcherStream goroutine stops.
	donec chan struct{}
}

// watchGrpcStream tracks all watch resources attached to a single grpc stream.
type watchGrpcStream struct {
	// wg is Done when all substream goroutines have exited
	wg sync.WaitGroup

	// ctx controls internal remote.Watch requests
	ctx context.Context

	remote   pb.WatchClient
	callOpts []grpc.CallOption

	// substreams holds all active watchers on this grpc stream
	substreams map[int64]*watcherStream
	// resuming holds all resuming watchers on this grpc stream
	resuming []*watcherStream
	// resumec closes to signal that all substreams should begin resuming
	resumec chan struct{}

	// requestChan sends a watch request from Watch() to the main goroutine
	requestChan chan watchStreamRequest
	// responseChan receives data from the watch client
	responseChan chan *pb.WatchResponse

	// donec closes to broadcast shutdown
	donec chan struct{}
	// errc transmits errors from grpc Recv to the watch stream reconnect logic
	errc chan error
	// closeErr is the error that closed the watch stream
	closeErr error
	// closingc gets the watcherStream of closing watchers
	closingc chan *watcherStream
}

// INFO: grpcStream 对象会 goroutine 调用 watch grpc server Watch()
func (w *watcher) newWatcherGrpcStream(ctx context.Context) *watchGrpcStream {
	grpcStream := &watchGrpcStream{
		ctx:    ctx,
		remote: w.remote,

		requestChan:  make(chan watchStreamRequest),
		responseChan: make(chan *pb.WatchResponse),
		donec:        make(chan struct{}),
		errc:         make(chan error),
		resumec:      make(chan struct{}),
	}

	go grpcStream.run()

	return grpcStream
}

// run is the root of the goroutines for managing a watcher client
func (grpcStream *watchGrpcStream) run() {
	var watchClient pb.Watch_WatchClient
	var closeErr error

	// start a stream with the etcd grpc server
	if watchClient, closeErr = grpcStream.newWatchClient(); closeErr != nil {
		return
	}

	cancelSet := make(map[int64]struct{})
	var cur *pb.WatchResponse
	for {
		select {
		// INFO: 提交请求到 watch grpc server
		case pbRequestChan := <-grpcStream.requestChan:
			switch wreq := pbRequestChan.(type) {
			case *watchRequest:
				klog.Infof(fmt.Sprintf("[run]watchRequest %s", wreq.toPB().String()))
				outc := make(chan WatchResponse, 1)
				ws := &watcherStream{
					initReq: *wreq,
					id:      -1,
					outc:    outc,
					// unbuffered so resumes won't cause repeat events
					receiveChan: make(chan *WatchResponse),
				}

				grpcStream.wg.Add(1)
				go grpcStream.serveSubstream(ws, grpcStream.resumec)

				// queue up for watcher creation/resume
				grpcStream.resuming = append(grpcStream.resuming, ws)
				if len(grpcStream.resuming) == 1 {
					// head of resume queue, can register a new watcher
					// INFO: send pbWatchRequest to watch grpc server
					if err := watchClient.Send(ws.initReq.toPB()); err != nil {
						klog.Errorf(fmt.Sprintf("[watchGrpcStream run]error when sending request: %v", err))
					}
				}
				//case *progressRequest:

			}

		// new events from the watch client
		// INFO: 从 watch grpc server 中获取 pbWatchResponse
		case pbWatchResponse := <-grpcStream.responseChan:
			klog.Infof(fmt.Sprintf("[run]pbWatchResponse %s", pbWatchResponse.String()))
			if cur == nil || pbWatchResponse.Created || pbWatchResponse.Canceled {
				cur = pbWatchResponse
			} else if cur != nil && cur.WatchId == pbWatchResponse.WatchId {
				// merge new events
				cur.Events = append(cur.Events, pbWatchResponse.Events...)
				// update "Fragment" field; last response with "Fragment" == false
				cur.Fragment = pbWatchResponse.Fragment
			}
			switch {
			case pbWatchResponse.Created:
				// response to head of queue creation
				if len(grpcStream.resuming) != 0 {
					if ws := grpcStream.resuming[0]; ws != nil {
						grpcStream.addSubstream(pbWatchResponse, ws)
						grpcStream.dispatchEvent(pbWatchResponse)
						grpcStream.resuming[0] = nil
					}
				}

				if ws := grpcStream.nextResume(); ws != nil {
					if err := watchClient.Send(ws.initReq.toPB()); err != nil {
						klog.Errorf(fmt.Sprintf("error when sending request: %v", err))
					}
				}

				// reset for next iteration
				cur = nil
			default:
				// dispatch to appropriate watch stream
				ok := grpcStream.dispatchEvent(cur)
				// reset for next iteration
				cur = nil
				if ok {
					break
				}
				// watch response on unexpected watch id; cancel id
				if _, ok := cancelSet[pbWatchResponse.WatchId]; ok {
					break
				}

				cancelSet[pbWatchResponse.WatchId] = struct{}{}
				cr := &pb.WatchRequest_CancelRequest{
					CancelRequest: &pb.WatchCancelRequest{
						WatchId: pbWatchResponse.WatchId,
					},
				}
				req := &pb.WatchRequest{RequestUnion: cr}
				klog.Infof(fmt.Sprintf("sending watch cancel request for failed dispatch watch-id: %d", pbWatchResponse.WatchId))
				if err := watchClient.Send(req); err != nil {
					klog.Errorf(fmt.Sprintf("failed to send watch cancel request watch-id: %d, err: %v", pbWatchResponse.WatchId, err))
				}
			}

		}
	}
}

// nextResume chooses the next resuming to register with the grpc stream. Abandoned
// streams are marked as nil in the queue since the head must wait for its inflight registration.
func (grpcStream *watchGrpcStream) nextResume() *watcherStream {
	for len(grpcStream.resuming) != 0 {
		if grpcStream.resuming[0] != nil {
			return grpcStream.resuming[0]
		}
		grpcStream.resuming = grpcStream.resuming[1:len(grpcStream.resuming)]
	}
	return nil
}

func (grpcStream *watchGrpcStream) addSubstream(resp *pb.WatchResponse, ws *watcherStream) {
	// check watch ID for backward compatibility (<= v3.3)
	if resp.WatchId == -1 || (resp.Canceled && resp.CancelReason != "") {
		grpcStream.closeErr = v3rpc.Error(errors.New(resp.CancelReason))
		// failed; no channel
		close(ws.receiveChan)
		return
	}
	ws.id = resp.WatchId
	grpcStream.substreams[ws.id] = ws
}

// dispatchEvent sends a WatchResponse to the appropriate watcher stream
func (grpcStream *watchGrpcStream) dispatchEvent(watchResponse *pb.WatchResponse) bool {
	events := make([]*Event, len(watchResponse.Events))
	for i, ev := range watchResponse.Events {
		events[i] = (*Event)(ev)
	}
	// TODO: return watch ID?
	wr := &WatchResponse{
		Header:          *watchResponse.Header,
		Events:          events,
		CompactRevision: watchResponse.CompactRevision,
		Created:         watchResponse.Created,
		Canceled:        watchResponse.Canceled,
		cancelReason:    watchResponse.CancelReason,
	}

	// watch IDs are zero indexed, so request notify watch responses are assigned a watch ID of -1 to
	// indicate they should be broadcast.
	if wr.IsProgressNotify() && watchResponse.WatchId == -1 {
		return grpcStream.broadcastResponse(wr)
	}

	return grpcStream.unicastResponse(wr, watchResponse.WatchId)
}

// broadcastResponse send a watch response to all watch substreams.
func (grpcStream *watchGrpcStream) broadcastResponse(wr *WatchResponse) bool {
	for _, ws := range grpcStream.substreams {
		select {
		case ws.receiveChan <- wr:
		case <-ws.donec:
		}
	}
	return true
}

// unicastResponse sends a watch response to a specific watch substream.
func (grpcStream *watchGrpcStream) unicastResponse(wr *WatchResponse, watchId int64) bool {
	ws, ok := grpcStream.substreams[watchId]
	if !ok {
		return false
	}
	select {
	case ws.receiveChan <- wr:
	case <-ws.donec:
		return false
	}
	return true
}

// INFO: 重点是 grpcStream.responseChan <- watchClient.Recv()
func (grpcStream *watchGrpcStream) newWatchClient() (pb.Watch_WatchClient, error) {
	// mark all substreams as resuming
	close(grpcStream.resumec)
	grpcStream.resumec = make(chan struct{})
	grpcStream.joinSubstreams()
	for _, ws := range grpcStream.substreams {
		ws.id = -1
		grpcStream.resuming = append(grpcStream.resuming, ws)
	}
	// strip out nils, if any
	var resuming []*watcherStream
	for _, ws := range grpcStream.resuming {
		if ws != nil {
			resuming = append(resuming, ws)
		}
	}
	grpcStream.resuming = resuming
	grpcStream.substreams = make(map[int64]*watcherStream)

	// connect to grpc stream while accepting watcher cancelation
	stopc := make(chan struct{})
	donec := grpcStream.waitCancelSubstreams(stopc)
	watchClient, err := grpcStream.openWatchClient()
	close(stopc)
	<-donec

	// serve all non-closing streams, even if there's a client error
	// so that the teardown path can shutdown the streams as expected.
	for _, ws := range grpcStream.resuming {
		if ws.closing {
			continue
		}
		ws.donec = make(chan struct{})
		grpcStream.wg.Add(1)
		go grpcStream.serveSubstream(ws, grpcStream.resumec)
	}
	if err != nil {
		return nil, v3rpc.Error(err)
	}

	// receive data from new grpc stream
	go grpcStream.serveWatchClient(watchClient)

	return watchClient, nil
}

// openWatchClient retries opening a watch client until success or halt.
// manually retry in case "ws==nil && err==nil"
// TODO: remove FailFast=false
func (grpcStream *watchGrpcStream) openWatchClient() (watchClient pb.Watch_WatchClient, err error) {
	for {
		select {
		case <-grpcStream.ctx.Done():
			if err == nil {
				return nil, grpcStream.ctx.Err()
			}
			return nil, err
		default:
		}

		// 这里通过 pb.WatchClient.Watch() 来 grpc 调用 watch grpc server
		if watchClient, err = grpcStream.remote.Watch(grpcStream.ctx, grpcStream.callOpts...); watchClient != nil && err == nil {
			break
		}

	}

	return watchClient, nil
}

// joinSubstreams waits for all substream goroutines to complete.
func (grpcStream *watchGrpcStream) joinSubstreams() {
	for _, ws := range grpcStream.substreams {
		<-ws.donec
	}
	for _, ws := range grpcStream.resuming {
		if ws != nil {
			<-ws.donec
		}
	}
}

// serveWatchClient forwards messages from the grpc stream to run()
func (grpcStream *watchGrpcStream) serveWatchClient(watchClient pb.Watch_WatchClient) {
	for {
		watchResponse, err := watchClient.Recv()
		if err != nil {
			select {
			case grpcStream.errc <- err:
			case <-grpcStream.donec:
			}
			return
		}
		select {
		case grpcStream.responseChan <- watchResponse:
		case <-grpcStream.donec:
			return
		}
	}
}

// serveSubstream forwards watch responses from run() to the subscriber
func (grpcStream *watchGrpcStream) serveSubstream(ws *watcherStream, resumec chan struct{}) {
	if ws.closing {
		panic("created substream goroutine but substream is closing")
	}

	// nextRev is the minimum expected next revision
	nextRev := ws.initReq.rev
	resuming := false
	defer func() {
		if !resuming {
			ws.closing = true
		}
		close(ws.donec)
		if !resuming {
			grpcStream.closingc <- ws
		}
		grpcStream.wg.Done()
	}()

	emptyWr := &WatchResponse{}
	for {
		curWr := emptyWr
		outc := ws.outc

		if len(ws.buf) > 0 {
			curWr = ws.buf[0]
		} else {
			outc = nil
		}

		select {
		case outc <- *curWr:
			if ws.buf[0].Err() != nil {
				return
			}
			ws.buf[0] = nil
			ws.buf = ws.buf[1:]
		case wr, ok := <-ws.receiveChan:
			if !ok {
				// shutdown from closeSubstream
				return
			}
			if len(wr.Events) != 0 {
				for _, event := range wr.Events {
					// {key:"hello" create_revision:2 mod_revision:3 version:2 value:"world"}
					klog.Infof(fmt.Sprintf("[serveSubstream]{%+v}", event.Kv.String()))
				}
			} else {
				klog.Infof(fmt.Sprintf("[serveSubstream]%+v", *wr))
			}
			if wr.Created {
				if ws.initReq.watchResponseChan != nil {
					ws.initReq.watchResponseChan <- ws.outc
					// to prevent next write from taking the slot in buffered channel
					// and posting duplicate create events
					ws.initReq.watchResponseChan = nil

					// send first creation event only if requested
					if ws.initReq.createdNotify {
						ws.outc <- *wr
					}
					if ws.initReq.rev == 0 {
						nextRev = wr.Header.Revision
					}
				}
			} else {
				// current progress of watch; <= store revision
				nextRev = wr.Header.Revision
			}
			if len(wr.Events) > 0 {
				nextRev = wr.Events[len(wr.Events)-1].Kv.ModRevision + 1
			}
			ws.initReq.rev = nextRev

			// created event is already sent above,
			// watcher should not post duplicate events
			if wr.Created {
				continue
			}

			// TODO pause channel if buffer gets too large
			ws.buf = append(ws.buf, wr)
		case <-resumec:
			resuming = true
			return
		}
	}
}

func (grpcStream *watchGrpcStream) waitCancelSubstreams(stopc <-chan struct{}) <-chan struct{} {
	var wg sync.WaitGroup
	wg.Add(len(grpcStream.resuming))
	donec := make(chan struct{})
	for i := range grpcStream.resuming {
		go func(ws *watcherStream) {
			defer wg.Done()
			if ws.closing {
				if ws.initReq.ctx.Err() != nil && ws.outc != nil {
					close(ws.outc)
					ws.outc = nil
				}
				return
			}
			select {
			case <-ws.initReq.ctx.Done():
				// closed ws will be removed from resuming
				ws.closing = true
				close(ws.outc)
				ws.outc = nil
				grpcStream.wg.Add(1)
				go func() {
					defer grpcStream.wg.Done()
					grpcStream.closingc <- ws
				}()
			case <-stopc:
			}
		}(grpcStream.resuming[i])
	}
	go func() {
		defer close(donec)
		wg.Wait()
	}()
	return donec
}
