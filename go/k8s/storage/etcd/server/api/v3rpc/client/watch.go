package client

import (
	"context"
	"fmt"
	"sync"

	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
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
		gRPCStream := w.streams[ctxKey]
		if gRPCStream == nil {
			gRPCStream = w.newWatcherGrpcStream(ctx)
			w.streams[ctxKey] = gRPCStream
		}
		donec := gRPCStream.donec
		reqc := gRPCStream.reqc
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

// watchGrpcStream tracks all watch resources attached to a single grpc stream.
type watchGrpcStream struct {
}

func (w *watcher) newWatcherGrpcStream(ctx context.Context) *watchGrpcStream {
	gRPCStream := &watchGrpcStream{}

	go gRPCStream.run()

	return gRPCStream
}

func (gRPCStream *watchGrpcStream) run() {

}
