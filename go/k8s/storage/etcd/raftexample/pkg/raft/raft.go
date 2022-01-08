package raft

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"k8s.io/klog/v2"
	"os"
	"strconv"
	"sync"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/etcd/server/v3/wal"
	"go.etcd.io/etcd/server/v3/wal/walpb"
)

// apply contains entries, snapshot to be applied. Once
// an apply is consumed, the entries will be persisted to
// to raft storage concurrently; the application must read
// raftDone before assuming the raft messages are stable.
type Apply struct {
	entries  []raftpb.Entry
	snapshot raftpb.Snapshot
	// notifyc synchronizes etcd server applies with the raft node
	notifyc chan struct{}
}

type Config struct {
	id    int // client ID for raft session
	peers []string
}

type RaftNode struct {
	tickMu *sync.Mutex

	id                       int      // client ID for raft session
	peers                    []string // raft peer URLs
	join                     bool     // node is joining an existing cluster
	snapdir                  string   // path to snapshot directory
	getSnapshotDataFromStore func() ([]byte, error)
	lastIndex                uint64 // index of log at start

	raft.Node
	wal         *wal.WAL
	waldir      string // path to WAL directory
	transport   *rafthttp.Transport
	raftStorage *raft.MemoryStorage
	storage     etcdserver.Storage // wal + snapshotter
	snapshotter *snap.Snapshotter
	confState   raftpb.ConfState

	// a chan to send out apply
	applyc  chan Apply
	commitC chan []string
	stopped chan struct{}

	logger *zap.Logger
}

func NewRaftNode(config *Config) *RaftNode {
	r := &RaftNode{
		id:      config.id,
		waldir:  fmt.Sprintf("data/wal-%d", config.id),
		snapdir: fmt.Sprintf("data/snap-%d", config.id),
		peers:   config.peers,

		applyc:  make(chan Apply),
		commitC: make(chan []string),
		stopped: make(chan struct{}),

		logger: zap.NewExample(),
	}

	return r
}

type raftReadyHandler struct {
	getLead              func() (lead uint64)
	updateLead           func(lead uint64)
	updateLeadership     func(newLeader bool)
	updateCommittedIndex func(uint64)
}

func (r *RaftNode) start(rh *raftReadyHandler) {
	// INFO: (1) replay wal from snapshot
	if !fileutil.Exist(r.snapdir) {
		if err := os.Mkdir(r.snapdir, 0750); err != nil {
			klog.Fatalf("raftexample: cannot create dir for snapshot (%v)", err)
		}
	}
	r.snapshotter = snap.New(zap.NewExample(), r.snapdir)
	var snapshot *raftpb.Snapshot
	var err error
	walExist := wal.Exist(r.waldir)
	if walExist {
		walSnaps, err := wal.ValidSnapshotEntries(r.logger, r.waldir)
		if err != nil {
			klog.Fatalf("raftexample: error listing snapshots (%v)", err)
		}
		snapshot, err = r.snapshotter.LoadNewestAvailable(walSnaps)
		if err != nil && err != snap.ErrNoSnapshot {
			klog.Fatalf("raftexample: error loading snapshot (%v)", err)
		}
	}
	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	if !walExist {
		if err := os.Mkdir(r.waldir, 0750); err != nil {
			klog.Fatalf("raftexample: cannot create dir for wal (%v)", err)
		}
		r.wal, err = wal.Create(r.logger, r.waldir, nil)
		if err != nil {
			klog.Fatalf("raftexample: create wal error (%v)", err)
		}
		r.wal.Close()
	}
	klog.Infof("loading WAL at term %d and index %d", walsnap.Term, walsnap.Index)
	r.wal, err = wal.Open(r.logger, r.waldir, walsnap)
	if err != nil {
		klog.Fatalf("raftexample: error loading wal (%v)", err)
	}
	_, state, ents, err := r.wal.ReadAll()
	if err != nil {
		klog.Fatalf("raftexample: failed to read WAL (%v)", err)
	}
	// append to storage so raft starts at the right place in log
	r.raftStorage = raft.NewMemoryStorage()
	if snapshot != nil {
		r.raftStorage.ApplySnapshot(*snapshot)
	}
	r.raftStorage.SetHardState(state)
	r.raftStorage.Append(ents) // INFO: wal 是 log entry 持久化存储，开始启动节点后，replay 到 memoryStorage 里

	r.storage = etcdserver.NewStorage(r.wal, r.snapshotter)

	// INFO: (2) start raft node
	peers := make([]raft.Peer, len(r.peers))
	for i := range peers {
		peers[i] = raft.Peer{ID: uint64(i + 1)}
	}
	c := &raft.Config{
		ID:                        uint64(r.id),
		ElectionTick:              10,
		HeartbeatTick:             1,
		Storage:                   r.raftStorage,
		MaxSizePerMsg:             1024 * 1024, // max byte size of each append message, 1MB
		MaxInflightMsgs:           256,
		MaxUncommittedEntriesSize: 1 << 30,
	}
	if walExist || r.join {
		r.Node = raft.RestartNode(c)
	} else {
		r.Node = raft.StartNode(c, peers)
	}

	// INFO: (3) start transport
	r.transport = &rafthttp.Transport{
		Logger:      zap.NewExample(),
		ID:          types.ID(r.id),
		ClusterID:   0x1000,
		Raft:        r,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(zap.NewExample(), strconv.Itoa(r.id)),
		ErrorC:      make(chan error),
	}
	r.transport.Start()
	for i := range r.peers {
		if i+1 != r.id {
			r.transport.AddPeer(types.ID(i+1), []string{r.peers[i]})
		}
	}

	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond) // 1s
		defer ticker.Stop()
		islead := false

		for {
			select {
			case <-ticker.C:
				// ElectionTick 50 * 1s，follower 没有收到 leader 的 msg，则开始 campaign;
				// HeartbeatTick 5 * 1s, leader 每秒给 follower 发心跳，不能超过 5s 还不发；
				r.Tick()

			case rd := <-r.Ready():
				if rd.SoftState != nil {
					newLeader := rd.SoftState.Lead != raft.None && rh.getLead() != rd.SoftState.Lead
					rh.updateLead(rd.SoftState.Lead)
					rh.updateLeadership(newLeader)

					islead = rd.RaftState == raft.StateLeader
				}

				notifyc := make(chan struct{}, 1)
				ap := Apply{
					entries:  rd.CommittedEntries,
					snapshot: rd.Snapshot,
					notifyc:  notifyc,
				}

				updateCommittedIndex(&ap, rh)

				// send out CommittedEntries/Snapshot
				select {
				case r.applyc <- ap:
				case <-r.stopped:
					return
				}

				data := r.publishEntries(rd.CommittedEntries)
				select {
				case r.commitC <- data:
				case <-r.stopped:
					return
				}

				// the leader can write to its disk in parallel with replicating to the followers and them
				// writing to their disks.
				// For more details, check raft thesis 10.2.1
				if islead {
					r.transport.Send(rd.Messages)
					//r.transport.Send(r.processMessages(rd.Messages))
				}

				// Must save the snapshot file and WAL snapshot entry before saving any other entries or hardstate to
				// ensure that recovery after a snapshot restore is possible.
				if !raft.IsEmptySnap(rd.Snapshot) {
					if err := r.storage.SaveSnap(rd.Snapshot); err != nil {
						r.logger.Fatal("failed to save Raft snapshot", zap.Error(err))
					}
				}
				if err := r.storage.Save(rd.HardState, rd.Entries); err != nil {
					r.logger.Fatal("failed to save Raft hard state and entries", zap.Error(err))
				}
				if !raft.IsEmptySnap(rd.Snapshot) {
					if err := r.storage.Sync(); err != nil {
						r.logger.Fatal("failed to sync Raft snapshot", zap.Error(err))
					}
					r.raftStorage.ApplySnapshot(rd.Snapshot)
					r.logger.Info("applied incoming Raft snapshot", zap.Uint64("snapshot-index", rd.Snapshot.Metadata.Index))
					if err := r.storage.Release(rd.Snapshot); err != nil {
						r.logger.Fatal("failed to release Raft wal", zap.Error(err))
					}
				}

				r.raftStorage.Append(rd.Entries)

				//if !islead {
				//	// finish processing incoming messages before we signal raftdone chan
				//	msgs := r.processMessages(rd.Messages)
				//
				//	// now unblocks 'applyAll' that waits on Raft log disk writes before triggering snapshots
				//	notifyc <- struct{}{}
				//
				//	// Candidate or follower needs to wait for all pending configuration
				//	// changes to be applied before sending messages.
				//	// Otherwise we might incorrectly count votes (e.g. votes from removed members).
				//	// Also slow machine's follower raft-layer could proceed to become the leader
				//	// on its own single-node cluster, before apply-layer applies the config change.
				//	// We simply wait for ALL pending entries to be applied for now.
				//	// We might improve this later on if it causes unnecessary long blocking issues.
				//	waitApply := false
				//	for _, ent := range rd.CommittedEntries {
				//		if ent.Type == raftpb.EntryConfChange {
				//			waitApply = true
				//			break
				//		}
				//	}
				//	if waitApply {
				//		// blocks until 'applyAll' calls 'applyWait.Trigger'
				//		// to be in sync with scheduled config-change job
				//		// (assume notifyc has cap of 1)
				//		select {
				//		case notifyc <- struct{}{}:
				//		case <-r.stopped:
				//			return
				//		}
				//	}
				//
				//	r.transport.Send(msgs)
				//} else {
				//	// leader already processed 'MsgSnap' and signaled
				//	//notifyc <- struct{}{}
				//}

				r.Advance()

			case <-r.stopped:
				return
			}
		}
	}()
}

func (r *RaftNode) publishEntries(ents []raftpb.Entry) []string {
	if len(ents) == 0 {
		return nil
	}

	data := make([]string, 0, len(ents))
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				// ignore empty messages
				break
			}
			data = append(data, string(ents[i].Data))
		case raftpb.EntryConfChange: // TODO: 暂时不管 ConfChange
			var cc raftpb.ConfChange
			cc.Unmarshal(ents[i].Data)
			r.confState = *r.ApplyConfChange(cc)
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					r.transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(r.id) {
					klog.Infof("I've been removed from the cluster! Shutting down.")
					return nil
				}
				r.transport.RemovePeer(types.ID(cc.NodeID))
			}
		}
	}

	return data
}

func (r *RaftNode) tick() {
	r.tickMu.Lock()
	r.Tick()
	r.tickMu.Unlock()
}

func (r *RaftNode) Apply() chan Apply {
	return r.applyc
}

func (r *RaftNode) Commit() chan []string {
	return r.commitC
}

func (r *RaftNode) Process(ctx context.Context, m raftpb.Message) error {
	return r.Step(ctx, m)
}
func (r *RaftNode) IsIDRemoved(id uint64) bool                           { return false }
func (r *RaftNode) ReportUnreachable(id uint64)                          {}
func (r *RaftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}

func updateCommittedIndex(ap *Apply, rh *raftReadyHandler) {
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
