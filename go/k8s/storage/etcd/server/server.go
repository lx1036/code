package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/etcd/pkg/notify"
	"k8s-lx1036/k8s/storage/etcd/raft"

	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/pkg/v3/idutil"
	"go.etcd.io/etcd/pkg/v3/wait"
	"k8s.io/klog/v2"
)

var (
	ErrStopped = errors.New("etcdserver: server stopped")
)

type EtcdServer struct {
	id       types.ID
	reqIDGen *idutil.Generator

	// INFO: raft
	raftNode *RaftNode // uses 64-bit atomics; keep 64-bit aligned.

	// INFO: linearizable read
	readMu sync.RWMutex
	// leaderChanged is used to notify the linearizable read loop to drop the old read requests.
	leaderChanged *notify.Notifier
	// read routine notifies etcd server that it waits for reading by sending an empty struct to readwaitC
	readwaitc chan struct{}
	// readNotifier is used to notify the read routine that it can process the request
	// when there is no error
	readNotifier *notifier
	// INFO: 等待本状态机去追赶 leader 状态机数据
	applyWait wait.WaitTime

	// INFO: Apply

	// stopping is closed by run goroutine on shutdown.
	stopping chan struct{}
}

func NewServer(config *raft.Config, peers []raft.Peer) *EtcdServer {
	clusterNodeID := 123
	server := &EtcdServer{
		id:       types.ID(clusterNodeID),
		reqIDGen: idutil.NewGenerator(uint16(clusterNodeID), time.Now()),

		leaderChanged: notify.NewNotifier(),
		readwaitc:     make(chan struct{}, 1),
		applyWait:     wait.NewTimeList(),

		stopping: make(chan struct{}, 1),
	}

	server.raftNode = newRaftNode(config, peers)

	return server
}

func (server *EtcdServer) run() {

	server.raftNode.start(rh)

}

// INFO: 线性一致性读,
//  https://time.geekbang.org/column/article/335932
//
func (server *EtcdServer) linearizableReadLoop() {
	for {
		requestId := server.reqIDGen.Next()
		leaderChangedNotifier := server.leaderChanged.Receive()
		select {
		case <-leaderChangedNotifier:
			continue
		case <-server.readwaitc: // INFO：给 readwaitc channel 发消息，来线性一致性读, 该 loop 会一直阻塞在这
		case <-server.stopping:
			return
		}

		nextnr := newNotifier()
		server.readMu.Lock()
		readNotifier := server.readNotifier
		server.readNotifier = nextnr
		server.readMu.Unlock()

		// INFO: 从 leader 获取该线性一致性读请求的最新索引 committed index, 非常重要!!!
		//  leader 状态机中最新索引是 confirmedIndex
		confirmedIndex, err := server.requestCurrentIndex(leaderChangedNotifier, requestId) // 会阻塞
		if isStopped(err) {
			return
		}
		if err != nil {
			readNotifier.notify(err)
			continue
		}

		// INFO: 当前raft node状态机中已提交索引 applied index 必须大于等于 leader 中的 committed index，才会去读本状态机数据；
		//  否则，必须等待本状态机去追赶 leader 状态机数据
		appliedIndex := server.getAppliedIndex()
		if appliedIndex < confirmedIndex {
			select {
			case <-server.applyWait.Wait(confirmedIndex):
			case <-server.stopping:
				return
			}
		}

		// unblock all l-reads requested at indices before confirmedIndex
		readNotifier.notify(nil)
	}
}

func isStopped(err error) bool {
	return err == raft.ErrStopped || err == ErrStopped
}

func (server *EtcdServer) requestCurrentIndex(leaderChangedNotifier <-chan struct{}, requestId uint64) (uint64, error) {
	err := server.sendReadIndex(requestId)
	if err != nil {
		return 0, err
	}

	for {
		select {
		case readState := <-server.raftNode.readStateChan:
			requestIdBytes := uint64ToBigEndianBytes(requestId)
			gotOwnResponse := bytes.Equal(readState.RequestCtx, requestIdBytes)
			if !gotOwnResponse {
				// a previous request might time out. now we should ignore the response of it and
				// continue waiting for the response of the current requests.
				responseId := uint64(0)
				if len(readState.RequestCtx) == 8 {
					responseId = binary.BigEndian.Uint64(readState.RequestCtx)
				}
				klog.Warningf(fmt.Sprintf("ignored out-of-date read index response; local node read indexes queueing up and waiting to be in sync with leader, sent-request-id: %d, received-request-id:%d",
					requestId, responseId))
				continue
			}

			return readState.Index, nil
		case <-server.stopping:
			return 0, ErrStopped
		}
	}

}

func (server *EtcdServer) sendReadIndex(requestIndex uint64) error {
	ctxToSend := uint64ToBigEndianBytes(requestIndex)
	cctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	err := server.raftNode.ReadIndex(cctx, ctxToSend)
	cancel()
	if err == raft.ErrStopped {
		return err
	}
	if err != nil {
		return err
	}

	return nil
}

func uint64ToBigEndianBytes(number uint64) []byte {
	byteResult := make([]byte, 8)
	binary.BigEndian.PutUint64(byteResult, number)
	return byteResult
}
