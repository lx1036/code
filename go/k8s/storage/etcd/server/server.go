package server

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"k8s-lx1036/k8s/storage/etcd/pkg/notify"
	"k8s-lx1036/k8s/storage/etcd/raft"

	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/pkg/v3/idutil"
	"go.etcd.io/etcd/pkg/v3/wait"
	"go.etcd.io/etcd/raft/v3/raftpb"
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

	appliedIndex   uint64 // must use atomic operations to access; keep 64-bit aligned.
	committedIndex uint64 // must use atomic operations to access; keep 64-bit aligned.
	term           uint64 // must use atomic operations to access; keep 64-bit aligned.

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
	stop     chan struct{}
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
		stop:     make(chan struct{}),
	}

	r := bootstrapFromWAL()
	server.raftNode = r.newRaftNode()

	return server
}

type etcdProgress struct {
	confState raftpb.ConfState
	snapi     uint64
	appliedt  uint64
	appliedi  uint64
}

type raftReadyHandler struct {
	getLead              func() (lead uint64)
	updateLead           func(lead uint64)
	updateLeadership     func(newLeader bool)
	updateCommittedIndex func(uint64)
}

func (server *EtcdServer) run() {
	snapshot, err := server.raftNode.raftStorage.Snapshot()
	if err != nil {
		klog.Fatalf(fmt.Sprintf("failed to get snapshot from Raft storage err: %v", err))
	}

	rh := &raftReadyHandler{
		getLead:          nil,
		updateLead:       nil,
		updateLeadership: nil,
		updateCommittedIndex: func(committedIndex uint64) {
			currentCommittedIndex := server.getCommittedIndex()
			if committedIndex > currentCommittedIndex {
				server.setCommittedIndex(committedIndex)
			}
		},
	}
	server.raftNode.start(rh)

	ep := etcdProgress{
		confState: snapshot.Metadata.ConfState,
		snapi:     snapshot.Metadata.Index,
		appliedt:  snapshot.Metadata.Term,
		appliedi:  snapshot.Metadata.Index,
	}

	defer func() {
		server.raftNode.stop()
	}()

	for {
		select {
		// INFO: 在 raft.go::start() 里会写这个 channel，获取已经 committed entries
		case ap := <-server.raftNode.apply():
			server.applyAll(&ep, &ap)
		case <-server.stop:
			return
		}
	}
}

func (server *EtcdServer) setTerm(v uint64) {
	atomic.StoreUint64(&server.term, v)
}

func (server *EtcdServer) getTerm() uint64 {
	return atomic.LoadUint64(&server.term)
}

func (server *EtcdServer) setCommittedIndex(v uint64) {
	atomic.StoreUint64(&server.committedIndex, v)
}

func (server *EtcdServer) getCommittedIndex() uint64 {
	return atomic.LoadUint64(&server.committedIndex)
}

func (server *EtcdServer) setAppliedIndex(v uint64) {
	atomic.StoreUint64(&server.appliedIndex, v)
}

func (server *EtcdServer) getAppliedIndex() uint64 {
	return atomic.LoadUint64(&server.appliedIndex)
}

func (server *EtcdServer) applyAll(ep *etcdProgress, apply *apply) {

	server.applyEntries(ep, apply)

	<-apply.notifyc
}

func (server *EtcdServer) applyEntries(ep *etcdProgress, apply *apply) {
	if len(apply.entries) == 0 {
		return
	}

	firsti := apply.entries[0].Index
	if firsti > ep.appliedi+1 {
		klog.Fatalf(fmt.Sprintf("[applyEntries]unexpected committed entry index, current-applied-index: %d, first-committed-entry-index: %d",
			ep.appliedi, firsti))
	}

	var ents []raftpb.Entry
	if ep.appliedi+1-firsti < uint64(len(apply.entries)) {
		ents = apply.entries[ep.appliedi+1-firsti:]
	}
	if len(ents) == 0 {
		return
	}

	server.apply(ents, &ep.confState)

}

func (server *EtcdServer) applyEntryNormal(entry *raftpb.Entry) {
	panic("not implemented")
}

// apply takes entries received from Raft (after it has been committed) and
// applies them to the current state of the EtcdServer.
func (server *EtcdServer) apply(entries []raftpb.Entry, confState *raftpb.ConfState) (appliedTerm uint64, appliedIndex uint64, shouldStop bool) {
	klog.Infof(fmt.Sprintf("[EtcdServer apply]Applying entries num-entries: %d", len(entries)))
	for index := range entries {
		entry := entries[index]
		klog.Infof(fmt.Sprintf("[EtcdServer apply]Applying entry {index:%d, term:%d, type:%s}", entry.Index, entry.Term, entry.Type.String()))

		switch entry.Type {
		case raftpb.EntryNormal:
			server.applyEntryNormal(&entry)
			server.setAppliedIndex(entry.Index)
			server.setTerm(entry.Term)

		case raftpb.EntryConfChange:

		default:
			klog.Fatalf(fmt.Sprintf("[EtcdServer apply]must be either EntryNormal or EntryConfChange, unknown entry type: %s", entry.Type.String()))
		}

		appliedIndex, appliedTerm = entry.Index, entry.Term
	}

	return appliedTerm, appliedIndex, shouldStop
}
