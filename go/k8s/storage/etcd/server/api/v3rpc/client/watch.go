package client

import (
	"context"
	"fmt"
	"sync"

	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	v3rpc "go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Event mvccpb.Event

type WatchChan <-chan WatchResponse

type Watcher interface {
	Watch(ctx context.Context, key string) WatchChan

	// RequestProgress requests a progress notify response be sent in all watch channels.
	RequestProgress(ctx context.Context) error

	// Close closes the watcher and cancels all watch requests.
	Close() error
}

type WatchResponse struct {
	Header pb.ResponseHeader
	Events []*Event
}

type watcher struct {
	// protects the grpc streams map
	sync.Mutex

	remote pb.WatchClient

	// streams holds all the active grpc streams keyed by ctx value.
	streams map[string]*watchGrpcStream
}

// watchRequest is issued by the subscriber to start a new watcher
type watchRequest struct {
	ctx context.Context
	key string
	end string
	rev int64

	// watchResponseChan receives a chan WatchResponse once the watcher is established
	watchResponseChan chan chan WatchResponse // INFO: 可以参考
}

func NewWatcher(conn *grpc.ClientConn) Watcher {
	return NewWatchFromWatchClient(pb.NewWatchClient(conn))
}

func NewWatchFromWatchClient(wc pb.WatchClient) Watcher {
	w := &watcher{
		remote:  wc,
		streams: make(map[string]*watchGrpcStream),
	}

	return w
}

func (w *watcher) Watch(ctx context.Context, key string) WatchChan {

	request := &watchRequest{
		ctx:               ctx,
		createdNotify:     ow.createdNotify,
		key:               key,
		progressNotify:    ow.progressNotify,
		filters:           filters,
		prevKV:            ow.prevKV,
		watchResponseChan: make(chan chan WatchResponse, 1),
	}
	ok := false
	ctxKey := streamKeyFromCtx(ctx)

	var closeCh chan WatchResponse
	for {
		// find or allocate appropriate grpc watch stream
		w.Lock()
		if w.streams == nil {
			// closed
			w.Unlock()
			ch := make(chan WatchResponse)
			close(ch)
			return ch // return closed chan
		}
		grpcStream := w.streams[ctxKey]
		if grpcStream == nil {
			grpcStream = w.newWatcherGrpcStream(ctx)
			w.streams[ctxKey] = grpcStream
		}
		donec := grpcStream.donec
		reqc := grpcStream.reqc
		w.Unlock()

		// couldn't create channel; return closed channel
		if closeCh == nil {
			closeCh = make(chan WatchResponse, 1)
		}

		// submit request
		select {
		case reqc <- request:
			ok = true

		}

		// receive channel
		if ok {
			select {
			case watchResponse := <-request.watchResponseChan:
				return watchResponse
			case <-ctx.Done():

			}
		}

		break
	}

	close(closeCh)
	return closeCh
}

func (w *watcher) RequestProgress(ctx context.Context) error {
	panic("implement me")
}

func (w *watcher) Close() error {
	panic("implement me")
}

func streamKeyFromCtx(ctx context.Context) string {
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		return fmt.Sprintf("%+v", md)
	}
	return ""
}

// watchStreamRequest is a union of the supported watch request operation types
type watchStreamRequest interface {
	toPB() *pb.WatchRequest
}

// watchGrpcStream tracks all watch resources attached to a single grpc stream.
type watchGrpcStream struct {
	// wg is Done when all substream goroutines have exited
	wg sync.WaitGroup

	remote pb.WatchClient

	// requestChan sends a watch request from Watch() to the main goroutine
	requestChan chan watchStreamRequest
	// responseChan receives data from the watch client
	responseChan chan *pb.WatchResponse
	// donec closes to broadcast shutdown
	donec chan struct{}
	// errc transmits errors from grpc Recv to the watch stream reconnect logic
	errc chan error
}

func (w *watcher) newWatcherGrpcStream(ctx context.Context) *watchGrpcStream {
	grpcStream := &watchGrpcStream{}

	go grpcStream.run()

	return grpcStream
}

// run is the root of the goroutines for managing a watcher client
func (grpcStream *watchGrpcStream) run() {
	var wc pb.Watch_WatchClient
	var closeErr error

	// start a stream with the etcd grpc server
	if wc, closeErr = grpcStream.newWatchClient(); closeErr != nil {
		return
	}

	for {
		select {

		case pbRequestChan := <-grpcStream.requestChan:
			switch pbRequestChan.(type) {
			case *watchRequest:

				grpcStream.wg.Add(1)
				go grpcStream.serveSubstream(ws, grpcStream.resumec)

			case *progressRequest:

			}

		// new events from the watch client
		case pbWatchResponse := <-grpcStream.responseChan:

			switch {
			case pbWatchResponse.Created:

			default:

			}

		}
	}

}

func (grpcStream *watchGrpcStream) newWatchClient() (pb.Watch_WatchClient, error) {

	watchClient, err := grpcStream.openWatchClient()

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

// watcherStream represents a registered watcher
type watcherStream struct {

	// receiveChan buffers watch responses before publishing
	receiveChan chan *WatchResponse
}

// serveSubstream forwards watch responses from run() to the subscriber
func (grpcStream *watchGrpcStream) serveSubstream(ws *watcherStream, resumec chan struct{}) {

	for {

		select {
		case wr, ok := <-ws.receiveChan:

		}
	}

}
