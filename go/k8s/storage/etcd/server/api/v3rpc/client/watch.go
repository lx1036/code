package client

import (
	"context"
	"fmt"
	"sync"

	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	v3rpc "go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

	// CompactRevision is the minimum revision the watcher may receive.
	CompactRevision int64

	// Created is used to indicate the creation of the watcher.
	Created bool

	Canceled     bool
	cancelReason string

	closeErr error
}

// IsProgressNotify returns true if the WatchResponse is progress notification.
func (wr *WatchResponse) IsProgressNotify() bool {
	return len(wr.Events) == 0 && !wr.Canceled && !wr.Created && wr.CompactRevision == 0 && wr.Header.Revision != 0
}

// Err is the error value if this WatchResponse holds an error.
func (wr *WatchResponse) Err() error {
	switch {
	case wr.closeErr != nil:
		return v3rpc.Error(wr.closeErr)
	case wr.CompactRevision != 0:
		return v3rpc.ErrCompacted
	case wr.Canceled:
		if len(wr.cancelReason) != 0 {
			return v3rpc.Error(status.Error(codes.FailedPrecondition, wr.cancelReason))
		}
		return v3rpc.ErrFutureRev
	}
	return nil
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

	// get the previous key-value pair before the event happens
	prevKV bool

	// send created notification event if this field is true
	createdNotify bool

	// watchResponseChan receives a chan WatchResponse once the watcher is established
	watchResponseChan chan chan WatchResponse // INFO: 可以参考
}

// toPB converts an internal watch request structure to its protobuf WatchRequest structure.
func (wr *watchRequest) toPB() *pb.WatchRequest {
	req := &pb.WatchCreateRequest{
		StartRevision: wr.rev,
		Key:           []byte(wr.key),
		RangeEnd:      []byte(wr.end),
		//ProgressNotify: wr.progressNotify,
		//Filters:        wr.filters,
		PrevKv: wr.prevKV,
		//Fragment:       wr.fragment,
	}
	cr := &pb.WatchRequest_CreateRequest{CreateRequest: req}
	return &pb.WatchRequest{RequestUnion: cr}
}

func NewWatcher(conn *grpc.ClientConn) Watcher {
	return NewWatchFromWatchClient(pb.NewWatchClient(conn))
}

func NewWatchFromWatchClient(watchClient pb.WatchClient) Watcher {
	w := &watcher{
		remote:  watchClient,
		streams: make(map[string]*watchGrpcStream),
	}

	return w
}

func (w *watcher) Watch(ctx context.Context, key string) WatchChan {
	request := &watchRequest{
		ctx:               ctx,
		key:               key,
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
		requestChan := grpcStream.requestChan
		w.Unlock()

		// couldn't create channel; return closed channel
		if closeCh == nil {
			closeCh = make(chan WatchResponse, 1)
		}

		// INFO: (1)submit request, 这种提交请求方式很特别!!!
		select {
		case requestChan <- request: // INFO: 肯定监听了 grpcStream.requestChan channel
			ok = true
		case <-request.ctx.Done():
			ok = false
		case <-donec:
			ok = false
			if grpcStream.closeErr != nil {
				closeCh <- WatchResponse{Canceled: true, closeErr: grpcStream.closeErr}
				break
			}
			// retry; may have dropped stream from no ctxs
			continue
		}

		// receive channel
		if ok {
			select {
			case watchResponse := <-request.watchResponseChan:
				return watchResponse
			case <-ctx.Done():
			case <-donec:
				if grpcStream.closeErr != nil {
					closeCh <- WatchResponse{Canceled: true, closeErr: grpcStream.closeErr}
					break
				}
				// retry; may have dropped stream from no ctxs
				continue
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
