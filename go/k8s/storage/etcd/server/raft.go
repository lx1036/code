package server

import (
	"fmt"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/etcd/raft"

	"go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
)

// INFO: 对 raft 包的封装，给 Server 调用

const (
	// The max throughput of etcd will not exceed 100MB/s (100K * 1KB value).
	// Assuming the RTT is around 10ms, 1MB max size is large enough.
	maxSizePerMsg = 1 * 1024 * 1024
	// Never overflow the rafthttp buffer, which is 4096.
	maxInflightMsgs = 4096 / 8
)

// apply contains entries, snapshot to be applied.
type apply struct {
	entries  []raftpb.Entry
	snapshot raftpb.Snapshot
	// notifyc synchronizes etcd server applies with the raft node
	notifyc chan struct{}
}

type RaftNodeConfig struct {
	raft.Node

	raftStorage *raft.MemoryStorage // INFO: 这个 storage 是用的 raft MemoryStorage，存在内存里，这个很重要!!!
	storage     Storage             // INFO: 这个 storage 是 wal 持久化，存在文件磁盘里，这个很重要!!!
}

type RaftNode struct {
	RaftNodeConfig

	// a chan to send out readState
	readStateChan chan raft.ReadState
	// a chan to send out apply
	applyChan chan apply

	ticker *time.Ticker
	tickMu *sync.Mutex

	stoppedChan chan struct{}
	doneChan    chan struct{}
}

func newRaftNode(config RaftNodeConfig) *RaftNode {
	//func newRaftNode(config *raft.Config, peers []raft.Peer) *RaftNode {
	//node := raft.StartNode(config, peers)
	raftNode := &RaftNode{
		RaftNodeConfig: config,

		//Node:        node,
		//raftStorage: raft.NewMemoryStorage(),

		readStateChan: make(chan raft.ReadState, 1),
		applyChan:     make(chan apply),

		ticker: time.NewTicker(time.Duration(config.HeartbeatTick) * time.Second),
		tickMu: new(sync.Mutex),

		stoppedChan: make(chan struct{}),
		doneChan:    make(chan struct{}),
	}

	return raftNode
}

func (raftNode *RaftNode) start(rh *raftReadyHandler) {
	internalTimeout := time.Second

	go func() {
		defer func() {
			close(raftNode.doneChan)
		}()

		for {
			select {
			case <-raftNode.ticker.C:
				raftNode.safeTick()

			case ready := <-raftNode.Ready(): // INFO: Ready() 里返回的是用户提交的 []pb.Entry

				if len(ready.ReadStates) != 0 {
					select {
					// INFO: raftNode run() loop 里会去写这个 channel
					case raftNode.readStateChan <- ready.ReadStates[len(ready.ReadStates)-1]:
					case <-time.After(internalTimeout):
						klog.Warningf(fmt.Sprintf("timed out sending read state timeout: %s", internalTimeout.String()))
					case <-raftNode.stoppedChan:
						return
					}
				}

				notifyc := make(chan struct{}, 1)
				ap := apply{
					entries:  ready.CommittedEntries,
					snapshot: ready.Snapshot,
					notifyc:  notifyc,
				}

				// INFO: 更新全局 CommittedIndex
				updateCommittedIndex(&ap, rh)

				select {
				case raftNode.applyChan <- ap:
				case <-raftNode.stoppedChan:
					return
				}

				// INFO: 持久化 []pb.Entry 到 WAL
				if err := raftNode.storage.Save(ready.HardState, ready.Entries); err != nil {
					klog.Fatalf(fmt.Sprintf("[EtcdServer raftNode start]failed to save Raft hard state and entries err: %v", err))
				}

				// INFO: 保存到 raft log 内存里
				raftNode.raftStorage.Append(ready.Entries)

				raftNode.Advance()
			case <-raftNode.stoppedChan:
				return
			}
		}
	}()
}

func (raftNode *RaftNode) safeTick() {
	raftNode.tickMu.Lock()
	raftNode.Tick()
	raftNode.tickMu.Unlock()
}

func (raftNode *RaftNode) apply() chan apply {
	return raftNode.applyChan
}

func (raftNode *RaftNode) stop() {
	raftNode.stoppedChan <- struct{}{}
	<-raftNode.doneChan
}

func updateCommittedIndex(ap *apply, rh *raftReadyHandler) {
	var ci uint64
	if len(ap.entries) != 0 {
		ci = ap.entries[len(ap.entries)-1].Index
	}
	if ap.snapshot.Metadata.Index > ci {
		ci = ap.snapshot.Metadata.Index
	}
	if ci != 0 {
		rh.updateCommittedIndex(ci)
	}
}
