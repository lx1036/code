package server

import (
	"fmt"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/etcd/raft"

	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
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

	heartbeat time.Duration

	raftStorage *raft.MemoryStorage // INFO: 这个 storage 是用的 raft MemoryStorage，存在内存里，这个很重要!!!
	storage     Storage             // INFO: 这个 storage 是 wal 持久化，存在文件磁盘里，这个很重要!!!

	// TCP 模块
	transport rafthttp.Transporter
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
	//node := raft.StartNode(config, peers)
	raftNode := &RaftNode{
		RaftNodeConfig: config,

		//Node:        node,
		//raftStorage: raft.NewMemoryStorage(),

		readStateChan: make(chan raft.ReadState, 1),
		applyChan:     make(chan apply),

		ticker: time.NewTicker(config.heartbeat), // 1000ms
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

		islead := false

		for {
			select {
			case <-raftNode.ticker.C:
				raftNode.safeTick()

			case ready := <-raftNode.Ready(): // INFO: Ready() 里返回的是用户提交的 []pb.Entry
				islead = ready.RaftState == raft.StateLeader

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

				// INFO: 只有 leader 可以并发异步：log entry 持久化写到 WAL；和 log replication 到 followers。非常重要!!!
				//  For more details, check raft thesis 10.2.1
				if islead && raftNode.transport != nil { // debug in local
					raftNode.transport.Send(raftNode.processMessages(ready.Messages))
				}

				// INFO: 在 WAL/Memory-RaftLog 保存 log entry 之前，先保存 WAL snapshot，这样可以保证 recovery restore from snapshot
				if !raft.IsEmptySnap(ready.Snapshot) {
					if err := raftNode.storage.SaveSnap(ready.Snapshot); err != nil {
						klog.Fatalf(fmt.Sprintf("[RaftNode start]failed to save Raft snapshot err:%v", err))
					}
				}

				// INFO: 持久化 []pb.Entry 到 WAL
				if err := raftNode.storage.Save(ready.HardState, ready.Entries); err != nil {
					klog.Fatalf(fmt.Sprintf("[EtcdServer raftNode start]failed to save Raft hard state and entries err: %v", err))
				}

				// INFO: 保存到 raft log 内存里
				raftNode.raftStorage.Append(ready.Entries)

				if !islead {
					// finish processing incoming messages before we signal raftdone chan
					msgs := raftNode.processMessages(ready.Messages)

					// now unblocks 'applyAll' that waits on Raft log disk writes before triggering snapshots
					notifyc <- struct{}{}

					// TODO:
					waitApply := false
					for _, ent := range ready.CommittedEntries {
						if ent.Type == raftpb.EntryConfChange {
							waitApply = true
							break
						}
					}
					if waitApply {
						// blocks until 'applyAll' calls 'applyWait.Trigger'
						// to be in sync with scheduled config-change job
						// (assume notifyc has cap of 1)
						select {
						case notifyc <- struct{}{}:
						case <-raftNode.stoppedChan:
							return
						}
					}

					if raftNode.transport != nil {
						raftNode.transport.Send(msgs)
					}
				} else {
					// leader already processed 'MsgSnap' and signaled
					notifyc <- struct{}{}
				}

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

// INFO: 获取已经 committed entries
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
