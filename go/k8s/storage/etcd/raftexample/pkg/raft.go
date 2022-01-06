package pkg

import (
	"context"
	"fmt"
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

	id                       int      // client ID for raft session
	peers                    []string // raft peer URLs
	join                     bool     // node is joining an existing cluster
	snapdir                  string   // path to snapshot directory
	getSnapshotDataFromStore func() ([]byte, error)
	lastIndex                uint64 // index of log at start

	// snapshot
	snapshotter      *snap.Snapshotter
	snapshotterReady chan *snap.Snapshotter // signals when snapshotter is ready
	confState        raftpb.ConfState
	snapshotIndex    uint64
	appliedIndex     uint64
	snapCount        uint64

	// raft backing for the commit/error channel
	node        raft.Node
	raftStorage *raft.MemoryStorage
	wal         *wal.WAL
	waldir      string // path to WAL directory
	transport   *rafthttp.Transport

	stopc     chan struct{} // signals proposal channel closed
	httpstopc chan struct{} // signals http server to shutdown

	logger *zap.Logger
}

func NewRaftNode(id int, peers []string, join bool, getSnapshotDataFromStore func() ([]byte, error), proposeC <-chan string,
	confChangeC <-chan raftpb.ConfChange) (<-chan *commit, <-chan error, <-chan *snap.Snapshotter) {
	commitC := make(chan *commit)
	errorC := make(chan error)

	node := &raftNode{
		id:                       id,
		proposeC:                 proposeC,
		confChangeC:              confChangeC,
		commitC:                  commitC,
		errorC:                   errorC,
		peers:                    peers,
		join:                     join,
		waldir:                   fmt.Sprintf("raft/wal-%d", id),
		snapdir:                  fmt.Sprintf("raft/snap-%d", id),
		getSnapshotDataFromStore: getSnapshotDataFromStore,
		snapCount:                defaultSnapshotCount,
		stopc:                    make(chan struct{}),
		httpstopc:                make(chan struct{}),

		snapshotterReady: make(chan *snap.Snapshotter, 1),

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
		ElectionTick:              50,
		HeartbeatTick:             5,
		Storage:                   r.raftStorage,
		MaxSizePerMsg:             1024 * 1024,
		MaxInflightMsgs:           256,
		MaxUncommittedEntriesSize: 1 << 30,
	}
	if walExist || r.join {
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
	go r.servePeers()

	// INFO: (5) proposeC -> raft; applyC <- raft
	go r.serveChannels()
}

// INFO: @see https://github.com/etcd-io/etcd/blob/v3.5.1/server/embed/etcd.go#L530-L588
//  https://github.com/etcd-io/etcd/blob/v3.5.1/server/etcdserver/api/etcdhttp/peer.go#L39-L76
//  tcp serve peers，peers 之间消息传递：MsgHeartbeat 等，可见 https://github.com/etcd-io/etcd/blob/v3.5.1/raft/raftpb/raft.pb.go#L70-L94
func (r *raftNode) servePeers() {
	u, err := url.Parse(r.peers[r.id-1])
	if err != nil {
		klog.Fatalf("raftexample: Failed parsing URL (%v)", err)
	}

	ln, err := newListenerWithStopC(u.Host, r.httpstopc)
	if err != nil {
		klog.Fatalf("raftexample: Failed to listen rafthttp (%v)", err)
	}

	err = (&http.Server{Handler: r.transport.Handler()}).Serve(ln) // block
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
		for r.proposeC != nil && r.confChangeC != nil {
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
	ticker := time.NewTicker(1000 * time.Millisecond) // 1s
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// ElectionTick 50 * 1s，follower 没有收到 leader 的 msg，则开始 campaign;
			// HeartbeatTick 5 * 1s, leader 每秒给 follower 发心跳，不能超过 5s 还不发；
			r.node.Tick()

			// store raft entries to wal, then publish over commit channel
		case rd := <-r.node.Ready():
			r.wal.Save(rd.HardState, rd.Entries)

			if !raft.IsEmptySnap(rd.Snapshot) { // process snap
				// save the snapshot file before writing the snapshot to the wal.
				// This makes it possible for the snapshot file to become orphaned, but prevents
				// a WAL snapshot entry from having no corresponding snapshot file.
				walSnap := walpb.Snapshot{
					Index:     rd.Snapshot.Metadata.Index,
					Term:      rd.Snapshot.Metadata.Term,
					ConfState: &rd.Snapshot.Metadata.ConfState,
				}
				r.snapshotter.SaveSnap(rd.Snapshot)
				r.wal.SaveSnapshot(walSnap)
				r.wal.ReleaseLockTo(rd.Snapshot.Metadata.Index)

				// raft log in memory storage
				r.raftStorage.ApplySnapshot(rd.Snapshot)

				klog.Infof("publishing snapshot at index %d", r.snapshotIndex)
				if rd.Snapshot.Metadata.Index <= r.appliedIndex {
					klog.Fatalf("snapshot index [%d] should > progress.appliedIndex [%d]", rd.Snapshot.Metadata.Index, r.appliedIndex)
				}
				r.commitC <- nil // trigger kvstore to load snapshot
				r.confState = rd.Snapshot.Metadata.ConfState
				r.snapshotIndex = rd.Snapshot.Metadata.Index
				r.appliedIndex = rd.Snapshot.Metadata.Index
				klog.Infof("finished publishing snapshot at index %d", r.snapshotIndex)
			}

			r.raftStorage.Append(rd.Entries)
			r.transport.Send(rd.Messages)
			applyDoneC, ok := r.publishEntries(r.entriesToApply(rd.CommittedEntries))
			if !ok {
				r.stop()
				return
			}
			r.maybeTriggerSnapshot(applyDoneC)
			r.node.Advance()

		case err := <-r.transport.ErrorC:
			klog.Errorf(fmt.Sprintf("[raft readyC]transport err:%v", err))
			r.stop()
			return

		case <-r.stopc:
			r.stop()
			return
		}
	}
}

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
			r.confState = *r.node.ApplyConfChange(cc)
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

var snapshotCatchUpEntriesN uint64 = 10000

func (r *raftNode) maybeTriggerSnapshot(applyDoneC <-chan struct{}) {
	if r.appliedIndex-r.snapshotIndex <= r.snapCount { // 如果 appliedIndex - snapshotIndex > 10000，则开始 trigger snapshot
		return
	}

	// wait until all committed entries are applied (or server is closed)
	if applyDoneC != nil {
		select {
		case <-applyDoneC:
		case <-r.stopc:
			return
		}
	}

	klog.Infof("start snapshot [applied index: %d | last snapshot index: %d]", r.appliedIndex, r.snapshotIndex)
	data, err := r.getSnapshotDataFromStore()
	if err != nil {
		klog.Fatal(err)
	}
	snapshot, err := r.raftStorage.CreateSnapshot(r.appliedIndex, &r.confState, data)
	if err != nil {
		panic(err)
	}
	walSnap := walpb.Snapshot{
		Index:     snapshot.Metadata.Index,
		Term:      snapshot.Metadata.Term,
		ConfState: &snapshot.Metadata.ConfState,
	}
	r.snapshotter.SaveSnap(snapshot)
	r.wal.SaveSnapshot(walSnap)
	r.wal.ReleaseLockTo(snapshot.Metadata.Index)

	compactIndex := uint64(1)
	if r.appliedIndex > snapshotCatchUpEntriesN {
		compactIndex = r.appliedIndex - snapshotCatchUpEntriesN
	}
	if err := r.raftStorage.Compact(compactIndex); err != nil {
		panic(err)
	}
	klog.Infof("compacted log at index %d", compactIndex)
	r.snapshotIndex = r.appliedIndex
}

// stop closes http, closes all channels, and stops raft.
func (r *raftNode) stop() {
	r.stopHTTP()
	close(r.commitC)
	close(r.errorC)
	r.node.Stop()
}

func (r *raftNode) stopHTTP() {
	r.transport.Stop()
	close(r.httpstopc)
}

func (r *raftNode) Process(ctx context.Context, m raftpb.Message) error {
	return r.node.Step(ctx, m)
}
func (r *raftNode) IsIDRemoved(id uint64) bool                           { return false }
func (r *raftNode) ReportUnreachable(id uint64)                          {}
func (r *raftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}
