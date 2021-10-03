package v3rpc

import (
	"context"
	"io"
	"sync"

	"k8s-lx1036/k8s/storage/etcd/storage/mvcc"

	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
)

type watchServer struct {
	clusterID int64
	memberID  int64

	maxRequestBytes int

	watchable mvcc.WatchableKV
}

func NewWatchServer() pb.WatchServer {
	server := &watchServer{
		clusterID: int64(1),
		memberID:  int64(1),

		maxRequestBytes: grpcOverheadBytes,

		watchable: mvcc.New(),
	}

	return server
}

type serverWatchStream struct {
	sync.WaitGroup
	sync.RWMutex

	watchStream mvcc.WatchStream

	gRPCStream pb.Watch_WatchServer

	// tracks the watchID that stream might need to send progress to
	progress map[mvcc.WatchID]bool

	// closec indicates the stream is closed.
	closec chan struct{}
}

// Watch INFO: 主要起两个 goroutine loop，一个是 send loop, 一个是 receive loop
func (server *watchServer) Watch(stream pb.Watch_WatchServer) (err error) {

	sws := serverWatchStream{
		gRPCStream: stream,
	}

	sws.Add(1)
	go func() {
		defer sws.Done()
		sws.sendLoop()
	}()

	go func() {
		rerr := sws.receiveLoop()
	}()

	// INFO: 可能存在 receive goroutine loop finishes before send goroutine loop
	select {
	case err = <-errc:
		if err == context.Canceled {
			err = rpctypes.ErrGRPCWatchCanceled
		}
		close(sws.ctrlStream)
	case <-stream.Context().Done():
		err = stream.Context().Err()
		if err == context.Canceled {
			err = rpctypes.ErrGRPCWatchCanceled
		}
	}

	sws.close()

	return err
}

func (sws *serverWatchStream) close() {

	sws.Wait()
}

func (sws *serverWatchStream) sendLoop() {

	for {
		select {
		case watchResponse, ok := <-sws.watchStream.Chan():
			if !ok {
				return
			}

			pbWatchResponse := &pb.WatchResponse{
				Header:          sws.newResponseHeader(watchResponse.Revision),
				WatchId:         int64(watchResponse.WatchID),
				Events:          events,
				CompactRevision: watchResponse.CompactRevision,
				Canceled:        canceled,
			}

			// 返回给客户端
			serr := sws.gRPCStream.Send(pbWatchResponse)
			if serr != nil {
				return
			}

			sws.Lock()
			if len(evs) > 0 && sws.progress[watchResponse.WatchID] {
				// elide next progress update if sent a key update
				sws.progress[watchResponse.WatchID] = false
			}
			sws.Unlock()
		case <-sws.closec:
			return
		}
	}
}

func (sws *serverWatchStream) receiveLoop() error {
	for {
		watchRequest, err := sws.gRPCStream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch watchRequest.RequestUnion.(type) {
		case *pb.WatchRequest_CreateRequest:

		case *pb.WatchRequest_CancelRequest:

		case *pb.WatchRequest_ProgressRequest:

		default:
			continue
		}
	}
}
