package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"k8s-lx1036/k8s/storage/etcd/raft"

	"k8s.io/klog/v2"
)

// INFO: 线性一致性读
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

func (server *EtcdServer) requestCurrentIndex(leaderChangedNotifier <-chan struct{}, requestId uint64) (uint64, error) {
	err := server.sendReadIndex(requestId)
	if err != nil {
		return 0, err
	}

	for {
		select {
		case readState := <-server.raftNode.readStateChan: // INFO: raftNode run() loop 里会去写这个 channel
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
