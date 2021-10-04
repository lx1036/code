package v3rpc

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	betesting "k8s-lx1036/k8s/storage/etcd/storage/backend/testing"
	"k8s-lx1036/k8s/storage/etcd/storage/mvcc"

	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"go.etcd.io/etcd/server/v3/lease"
	"k8s.io/klog/v2"
)

type watchServer struct {
	clusterID int64
	memberID  int64

	tmpPath string // tmp db file

	maxRequestBytes int

	watchable mvcc.WatchableKV
}

func NewWatchServer() pb.WatchServer {
	b, tmpPath := betesting.NewDefaultTmpBackend()

	server := &watchServer{
		clusterID: int64(1),
		memberID:  int64(1),
		tmpPath:   tmpPath,

		maxRequestBytes: grpcOverheadBytes,

		watchable: mvcc.New(b, &lease.FakeLessor{}, mvcc.StoreConfig{}),
	}

	return server
}

type watchServerStream struct {
	sync.WaitGroup
	sync.RWMutex

	clusterID int64
	memberID  int64
	tmpPath   string // tmp db file

	pbWatchServer pb.Watch_WatchServer

	watchStream mvcc.WatchStream
	watchable   mvcc.WatchableKV

	pbWatchResponseChan chan *pb.WatchResponse

	// tracks the watchID that stream might need to send progress to
	progress map[mvcc.WatchID]bool
	// record watch IDs that need return previous key-value pair
	prevKV map[mvcc.WatchID]bool

	// closec indicates the stream is closed.
	closec chan struct{}
}

const ctrlStreamBufLen = 16

// Watch INFO: 主要起两个 goroutine loop，一个是 send loop, 一个是 receive loop
func (server *watchServer) Watch(pbWatchServer pb.Watch_WatchServer) (err error) {
	stream := watchServerStream{

		clusterID: server.clusterID,
		memberID:  server.memberID,
		tmpPath:   server.tmpPath,

		pbWatchServer: pbWatchServer,

		// INFO: 调用 mvcc 模块 WatchStream
		watchStream: server.watchable.NewWatchStream(),
		watchable:   server.watchable,

		// chan for sending control response like watcher created and canceled.
		pbWatchResponseChan: make(chan *pb.WatchResponse, ctrlStreamBufLen),
	}

	stream.Add(1)
	go func() {
		defer stream.Done()
		stream.sendLoop()
	}()

	errc := make(chan error, 1)
	go func() {
		if rerr := stream.receiveLoop(); rerr != nil {
			errc <- rerr
		}
	}()

	// INFO: 可能存在 receive goroutine loop finishes before send goroutine loop
	select {
	case err = <-errc:
		if err == context.Canceled {
			err = rpctypes.ErrGRPCWatchCanceled
		}
		close(stream.pbWatchResponseChan)
	case <-pbWatchServer.Context().Done():
		err = pbWatchServer.Context().Err()
		if err == context.Canceled {
			err = rpctypes.ErrGRPCWatchCanceled
		}
	}

	stream.close() // block
	return err
}

func (stream *watchServerStream) close() {
	defer os.RemoveAll(stream.tmpPath) // remove tmp db file
	stream.watchStream.Close()
	close(stream.closec)

	stream.Wait()
}

func (stream *watchServerStream) sendLoop() {
	for {
		select {
		case watchResponse, ok := <-stream.watchStream.Chan():
			if !ok {
				return
			}

			watchResponseEvents := watchResponse.Events
			events := make([]*mvccpb.Event, len(watchResponseEvents))
			stream.RLock()
			needPrevKV := stream.prevKV[watchResponse.WatchID]
			stream.RUnlock()
			for i := range watchResponseEvents {
				events[i] = &watchResponseEvents[i]
				// fill PrevKv
				if needPrevKV && !IsCreateEvent(watchResponseEvents[i]) {
					opt := mvcc.RangeOptions{Rev: watchResponseEvents[i].Kv.ModRevision - 1} // ModRevision-1 就是 prevKV
					r, err := stream.watchable.Range(context.TODO(), watchResponseEvents[i].Kv.Key, nil, opt)
					if err == nil && len(r.KVs) != 0 {
						events[i].PrevKv = &(r.KVs[0])
					}
				}
			}

			canceled := watchResponse.CompactRevision != 0
			pbWatchResponse := &pb.WatchResponse{
				Header:          stream.newResponseHeader(watchResponse.Revision),
				WatchId:         int64(watchResponse.WatchID),
				Events:          events,
				CompactRevision: watchResponse.CompactRevision,
				Canceled:        canceled,
			}

			// INFO: 返回给客户端，包含有 []mvccpb.Event 数据
			err := stream.pbWatchServer.Send(pbWatchResponse)
			if err != nil {
				klog.Errorf(fmt.Sprintf("[sendLoop]failed to send watch control response to gRPC stream err: %v", err))
				return
			}

			stream.Lock()
			if len(watchResponseEvents) > 0 && stream.progress[watchResponse.WatchID] {
				// elide next progress update if sent a key update
				stream.progress[watchResponse.WatchID] = false
			}
			stream.Unlock()
		case pbWatchResponse, ok := <-stream.pbWatchResponseChan: // 这个没有 Events 数据
			if !ok {
				return
			}

			if err := stream.pbWatchServer.Send(pbWatchResponse); err != nil {
				klog.Errorf(fmt.Sprintf("[sendLoop]failed to send watch control response to gRPC stream err: %v", err))
				return
			}
		case <-stream.closec:
			return
		}
	}
}

// INFO: 从 grpc watch client 接收到 watch request 对象
func (stream *watchServerStream) receiveLoop() error {
	for {
		watchRequest, err := stream.pbWatchServer.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch request := watchRequest.RequestUnion.(type) {
		case *pb.WatchRequest_CreateRequest:
			if request.CreateRequest == nil {
				break
			}

			createRequest := request.CreateRequest
			if len(createRequest.RangeEnd) == 0 {
				createRequest.RangeEnd = nil
			}

			kvStoreCurrentRevision := stream.watchStream.Rev() // KV store 当前 revision
			rev := createRequest.StartRevision
			if rev == 0 {
				rev = kvStoreCurrentRevision + 1
			}
			id, err := stream.watchStream.Watch(mvcc.WatchID(createRequest.WatchId), createRequest.Key, createRequest.RangeEnd, rev)
			if err == nil {
				stream.Lock()

				if createRequest.PrevKv {
					stream.prevKV[id] = true
				}

				stream.Unlock()
			}
			pbWatchResponse := &pb.WatchResponse{
				Header:   stream.newResponseHeader(kvStoreCurrentRevision),
				WatchId:  int64(id),
				Created:  true,
				Canceled: err != nil,
			}
			if err != nil {
				pbWatchResponse.CancelReason = err.Error()
			}

			select {
			case stream.pbWatchResponseChan <- pbWatchResponse:
			case <-stream.closec:
				return nil
			}
		case *pb.WatchRequest_CancelRequest:

		case *pb.WatchRequest_ProgressRequest:

		default:
			continue
		}
	}
}

const CurrentTerm = 10

func (stream *watchServerStream) newResponseHeader(rev int64) *pb.ResponseHeader {
	return &pb.ResponseHeader{
		ClusterId: uint64(stream.clusterID),
		MemberId:  uint64(stream.memberID),
		Revision:  rev,
		RaftTerm:  CurrentTerm,
	}
}

func IsCreateEvent(e mvccpb.Event) bool {
	return e.Type == mvccpb.PUT && e.Kv.CreateRevision == e.Kv.ModRevision
}
