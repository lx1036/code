package pkg

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/etcd/server/v3/wal"
	"go.etcd.io/etcd/server/v3/wal/walpb"
	"go.uber.org/zap"
	"k8s.io/klog/v2"
)

const (
	defaultSnapshotCount uint64 = 10000
)

// A key-value stream backed by raft
type raftNode struct {
	proposeC    <-chan string            // proposed messages (k,v)
	confChangeC <-chan raftpb.ConfChange // proposed cluster config changes
	commitC     chan<- *commit           // entries committed to log (k,v)
	errorC      chan<- error             // errors from raft session

	id          int      // client ID for raft session
	peers       []string // raft peer URLs
	join        bool     // node is joining an existing cluster
	snapdir     string   // path to snapshot directory
	getSnapshot func() ([]byte, error)
	lastIndex   uint64 // index of log at start

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

	// raft backing for the commit/error channel
	node        raft.Node
	raftStorage *raft.MemoryStorage
	wal         *wal.WAL
	waldir      string // path to WAL directory

	snapshotter      *snap.Snapshotter
	snapshotterReady chan *snap.Snapshotter // signals when snapshotter is ready

	snapCount uint64
	transport *rafthttp.Transport
	stopc     chan struct{} // signals proposal channel closed
	httpstopc chan struct{} // signals http server to shutdown
	httpdonec chan struct{} // signals http server shutdown complete

	logger *zap.Logger
}

func NewRaftNode(id int, peers []string, join bool, getSnapshot func() ([]byte, error), proposeC <-chan string,
	confChangeC <-chan raftpb.ConfChange) (<-chan *commit, <-chan error, <-chan *snap.Snapshotter) {
	commitC := make(chan *commit)
	errorC := make(chan error)

	node := &raftNode{
		id:          id,
		proposeC:    proposeC,
		confChangeC: confChangeC,
		commitC:     commitC,
		errorC:      errorC,
		peers:       peers,
		join:        join,
		waldir:      fmt.Sprintf("raftexample-%d", id),
		snapdir:     fmt.Sprintf("raftexample-%d-snap", id),
		getSnapshot: getSnapshot,
		snapCount:   defaultSnapshotCount,
		stopc:       make(chan struct{}),
		httpstopc:   make(chan struct{}),
		httpdonec:   make(chan struct{}),

		snapshotterReady: make(chan *snap.Snapshotter, 1),
		// rest of structure populated after WAL replay

		logger: zap.NewExample(),
	}

	go node.startRaft()

	return commitC, errorC, node.snapshotterReady
}

func (r *raftNode) startRaft() {
	// INFO: (1) replay wal from snapshot
	if !fileutil.Exist(r.snapdir) {
		if err := os.Mkdir(r.snapdir, 0750); err != nil {
			klog.Fatalf("raftexample: cannot create dir for snapshot (%v)", err)
		}
	}
	r.snapshotter = snap.New(zap.NewExample(), r.snapdir)
	oldWal := wal.Exist(r.waldir)
	klog.Infof("replaying WAL of member %d", r.id)
	snapshot, err := r.snapshotter.Load()
	if err != nil && err != snap.ErrNoSnapshot {
		klog.Fatalf("raftexample: error loading snapshot (%v)", err)
	}
	if !wal.Exist(r.waldir) {
		if err := os.Mkdir(r.waldir, 0750); err != nil {
			klog.Fatalf("raftexample: cannot create dir for wal (%v)", err)
		}
		r.wal, err = wal.Create(r.logger, r.waldir, nil)
		if err != nil {
			klog.Fatalf("raftexample: create wal error (%v)", err)
		}
		r.wal.Close()
	}
	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	klog.Infof("loading WAL at term %d and index %d", walsnap.Term, walsnap.Index)
	r.wal, err = wal.Open(r.logger, r.waldir, walsnap)
	if err != nil {
		klog.Fatalf("raftexample: error loading wal (%v)", err)
	}
	_, st, ents, err := r.wal.ReadAll()
	if err != nil {
		klog.Fatalf("raftexample: failed to read WAL (%v)", err)
	}
	r.raftStorage = raft.NewMemoryStorage()
	if snapshot != nil {
		r.raftStorage.ApplySnapshot(*snapshot)
	}
	r.raftStorage.SetHardState(st)
	// append to storage so raft starts at the right place in log
	r.raftStorage.Append(ents)
	r.snapshotterReady <- r.snapshotter

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
		MaxSizePerMsg:             1024 * 1024,
		MaxInflightMsgs:           256,
		MaxUncommittedEntriesSize: 1 << 30,
	}
	if oldWal || r.join {
		r.node = raft.RestartNode(c)
	} else {
		r.node = raft.StartNode(c, peers)
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

	// INFO: (4) start raft server
	go r.serveRaft()

	// INFO: (5) proposeC -> raft; applyC <- raft
	go r.serveChannels()
}

func (r *raftNode) serveRaft() {
	u, err := url.Parse(r.peers[r.id-1])
	if err != nil {
		klog.Fatalf("raftexample: Failed parsing URL (%v)", err)
	}

	ln, err := net.Listen("tcp", u.Host)
	if err != nil {
		klog.Fatalf("raftexample: Failed to listen rafthttp (%v)", err)
	}

	err = (&http.Server{Handler: r.transport.Handler()}).Serve(ln)
	if err != nil {
		klog.Error(err)
	}
}

func (r *raftNode) serveChannels() {
	snapshot, err := r.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	r.confState = snapshot.Metadata.ConfState
	r.snapshotIndex = snapshot.Metadata.Index
	r.appliedIndex = snapshot.Metadata.Index

	defer r.wal.Close()

	// INFO: 提交 proposeC 到 raft 里
	go func() {
		confChangeCount := uint64(0)
		for {
			select {
			case propose := <-r.proposeC:
				err := r.node.Propose(context.TODO(), []byte(propose))
				if err != nil {
					klog.Errorf(fmt.Sprintf("[raft proposeC]propose err:%v", err))
				}
				return

			case confChange := <-r.confChangeC:
				confChangeCount++
				confChange.ID = confChangeCount
				r.node.ProposeConfChange(context.TODO(), confChange)

			case <-r.stopc:
				r.stop()
				return
			}
		}
	}()

	// INFO: 从 raft 里读取 readyC
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.node.Tick()

			// store raft entries to wal, then publish over commit channel
		case rd := <-r.node.Ready():
			r.wal.Save(rd.HardState, rd.Entries)
			/*if !raft.IsEmptySnap(rd.Snapshot) {
				r.saveSnap(rd.Snapshot)
				r.raftStorage.ApplySnapshot(rd.Snapshot)
				r.publishSnapshot(rd.Snapshot)
			}*/
			r.raftStorage.Append(rd.Entries)
			r.transport.Send(rd.Messages)
			if _, ok := r.publishEntries(r.entriesToApply(rd.CommittedEntries)); !ok {
				r.stop()
				return
			}
			//r.maybeTriggerSnapshot(applyDoneC)
			r.node.Advance()

		case err := <-r.transport.ErrorC:
			klog.Errorf(fmt.Sprintf("[raft readyC]transport err:%v", err))
			return

		case <-r.stopc:
			r.stop()
			return
		}
	}
}

// TODO: 没理解？？？
func (r *raftNode) entriesToApply(ents []raftpb.Entry) (nents []raftpb.Entry) {
	if len(ents) == 0 {
		return ents
	}
	firstIdx := ents[0].Index
	if firstIdx > r.appliedIndex+1 {
		klog.Fatalf(fmt.Sprintf("first index of committed entry[%d] should <= progress.appliedIndex[%d]+1", firstIdx, r.appliedIndex))
	}
	if r.appliedIndex-firstIdx+1 < uint64(len(ents)) {
		nents = ents[r.appliedIndex-firstIdx+1:]
	}
	return nents
}

type commit struct {
	data       []string
	applyDoneC chan<- struct{}
}

// publishEntries writes committed log entries to commit channel and returns
// whether all entries could be published.
func (r *raftNode) publishEntries(ents []raftpb.Entry) (<-chan struct{}, bool) {
	if len(ents) == 0 {
		return nil, true
	}

	data := make([]string, 0, len(ents))
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				// ignore empty messages
				break
			}
			s := string(ents[i].Data)
			data = append(data, s)
		case raftpb.EntryConfChange: // TODO: 暂时不管 ConfChange
			var cc raftpb.ConfChange
			cc.Unmarshal(ents[i].Data)
			//r.confState = *r.node.ApplyConfChange(cc)
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					r.transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(r.id) {
					klog.Infof("I've been removed from the cluster! Shutting down.")
					return nil, false
				}
				r.transport.RemovePeer(types.ID(cc.NodeID))
			}
		}
	}

	var applyDoneC chan struct{}

	if len(data) > 0 {
		applyDoneC = make(chan struct{}, 1)
		select {
		case r.commitC <- &commit{data, applyDoneC}:
		case <-r.stopc:
			return nil, false
		}
	}

	// after commit, update appliedIndex
	r.appliedIndex = ents[len(ents)-1].Index

	return applyDoneC, true
}

func (r *raftNode) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	klog.Infof("publishing snapshot at index %d", r.snapshotIndex)
	defer klog.Infof("finished publishing snapshot at index %d", r.snapshotIndex)

	if snapshotToSave.Metadata.Index <= r.appliedIndex {
		klog.Fatalf("snapshot index [%d] should > progress.appliedIndex [%d]", snapshotToSave.Metadata.Index, r.appliedIndex)
	}

	r.commitC <- nil // trigger kvstore to load snapshot

	r.confState = snapshotToSave.Metadata.ConfState
	r.snapshotIndex = snapshotToSave.Metadata.Index
	r.appliedIndex = snapshotToSave.Metadata.Index
}

/*func (r *raftNode) saveSnap(snap raftpb.Snapshot) error {
	// must save the snapshot index to the WAL before saving the
	// snapshot to maintain the invariant that we only Open the
	// wal at previously-saved snapshot indexes.
	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}
	if err := r.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}
	if err := r.snapshotter.SaveSnap(snap); err != nil {
		return err
	}

	return r.wal.ReleaseLockTo(snap.Metadata.Index)
}*/

// stop closes http, closes all channels, and stops raft.
func (r *raftNode) stop() {
	r.stopHTTP()
	close(r.commitC)
	close(r.errorC)
	r.node.Stop()
}

func (r *raftNode) stopHTTP() {

}

func (r *raftNode) Process(ctx context.Context, m raftpb.Message) error {
	return r.node.Step(ctx, m)
}
func (r *raftNode) IsIDRemoved(id uint64) bool                           { return false }
func (r *raftNode) ReportUnreachable(id uint64)                          {}
func (r *raftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}
