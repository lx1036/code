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

// apply contains entries, snapshot to be applied.
type apply struct {
	entries  []raftpb.Entry
	snapshot raftpb.Snapshot
	// notifyc synchronizes etcd server applies with the raft node
	notifyc chan struct{}
}

type RaftNode struct {
	raft.Node

	raftStorage *raft.MemoryStorage

	// a chan to send out readState
	readStateChan chan raft.ReadState
	// a chan to send out apply
	applyChan chan apply
	
	ticker *time.Ticker
	tickMu *sync.Mutex

	stoppedChan chan struct{}
	doneChan    chan struct{}
}

func newRaftNode(config *raft.Config, peers []raft.Peer) *RaftNode {
	node := raft.StartNode(config, peers)
	raftNode := &RaftNode{
		Node:        node,
		raftStorage: raft.NewMemoryStorage(),

		readStateChan: make(chan raft.ReadState, 1),
		applyChan:     make(chan apply),
		
		ticker: time.NewTicker(time.Duration(int64(config.HeartbeatTick) * int64(time.Second))),
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
			
			case ready := <-raftNode.Ready(): // INFO:

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

				raftNode.raftStorage.Append(ready.Entries)

				raftNode.Advance()
			case <-raftNode.stoppedChan:
				return
			}
		}
	}()
}

func (raftNode *RaftNode) safeTick()  {
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
