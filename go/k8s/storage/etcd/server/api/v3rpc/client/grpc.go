package client

import (
	"context"
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

	// outc publishes watch responses to subscriber
	outc chan WatchResponse
	// recvc buffers watch responses before publishing
	recvc chan *WatchResponse

	// receiveChan buffers watch responses before publishing
	receiveChan chan *WatchResponse
}

// watchGrpcStream tracks all watch resources attached to a single grpc stream.
type watchGrpcStream struct {
	// wg is Done when all substream goroutines have exited
	wg sync.WaitGroup

	// ctx controls internal remote.Watch requests
	ctx context.Context

	remote   pb.WatchClient
	callOpts []grpc.CallOption

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
}

// grpcStream 对象会 goroutine 调用 watch grpc server Watch()
func (w *watcher) newWatcherGrpcStream(ctx context.Context) *watchGrpcStream {
	grpcStream := &watchGrpcStream{
		ctx:    ctx,
		remote: w.remote,

		requestChan:  make(chan watchStreamRequest),
		responseChan: make(chan *pb.WatchResponse),
		donec:        make(chan struct{}),
		errc:         make(chan error),
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
				outc := make(chan WatchResponse, 1)
				ws := &watcherStream{
					initReq: *wreq,
					id:      -1,
					outc:    outc,
					// unbuffered so resumes won't cause repeat events
					recvc: make(chan *WatchResponse),
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

// dispatchEvent sends a WatchResponse to the appropriate watcher stream
func (grpcStream *watchGrpcStream) dispatchEvent(pbresp *pb.WatchResponse) bool {

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

	for {

		select {
		case wr, ok := <-ws.receiveChan:

		}
	}

}
