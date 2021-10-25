package pkg

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"k8s-lx1036/k8s/storage/etcd/raft"
	"k8s-lx1036/k8s/storage/etcd/storage/wal"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/client/pkg/v3/types"
	etcdRaft "go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/etcd/server/v3/wal/walpb"
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
}

// NewRaftNode initiates a raft instance and returns a committed log entry
// channel and error channel. Proposals for log updates are sent over the
// provided the proposal channel. All log entries are replayed over the
// commit channel, followed by a nil message (to indicate the channel is
// current), then new log entries. To shutdown, close proposeC and read errorC.
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
	}

	go node.startRaft()

	return commitC, errorC, node.snapshotterReady
}

func (rNode *raftNode) startRaft() {
	if !fileutil.Exist(rNode.snapdir) {
		if err := os.Mkdir(rNode.snapdir, 0750); err != nil {
			klog.Fatalf("raftexample: cannot create dir for snapshot (%v)", err)
		}
	}

	rNode.snapshotter = snap.New(zap.NewExample(), rNode.snapdir)
	rNode.snapshotterReady <- rNode.snapshotter

	oldwal := wal.Exist(rNode.waldir)
	rNode.wal = rNode.replayWAL()

	peers := make([]raft.Peer, len(rNode.peers))
	for i := range peers {
		peers[i] = raft.Peer{ID: uint64(i + 1)}
	}
	c := &raft.Config{
		ID:                        uint64(rNode.id),
		ElectionTick:              10,
		HeartbeatTick:             1,
		Storage:                   rNode.raftStorage,
		MaxSizePerMsg:             1024 * 1024,
		MaxInflightMsgs:           256,
		MaxUncommittedEntriesSize: 1 << 30,
	}

	if oldwal || rNode.join {
		rNode.node = raft.RestartNode(c)
	} else {
		rNode.node = raft.StartNode(c, peers)
	}

	rNode.transport = &rafthttp.Transport{
		Logger:      zap.NewExample(),
		ID:          types.ID(rNode.id),
		ClusterID:   0x1000,
		Raft:        rNode,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(zap.NewExample(), strconv.Itoa(rNode.id)),
		ErrorC:      make(chan error),
	}
	rNode.transport.Start()
	for i := range rNode.peers {
		if i+1 != rNode.id {
			rNode.transport.AddPeer(types.ID(i+1), []string{rNode.peers[i]})
		}
	}

	go rNode.serveRaft()
	go rNode.serveChannels()
}

func (rNode *raftNode) serveRaft() {
	u, err := url.Parse(rNode.peers[rNode.id-1])
	if err != nil {
		klog.Fatalf("raftexample: Failed parsing URL (%v)", err)
	}

	ln, err := net.Listen("tcp", u.Host)
	if err != nil {
		klog.Fatalf("raftexample: Failed to listen rafthttp (%v)", err)
	}

	err = (&http.Server{Handler: rNode.transport.Handler()}).Serve(ln)
	if err != nil {
		klog.Error(err)
	}
}

func (rNode *raftNode) serveChannels() {
	snapshot, err := rNode.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	rNode.confState = snapshot.Metadata.ConfState
	rNode.snapshotIndex = snapshot.Metadata.Index
	rNode.appliedIndex = snapshot.Metadata.Index

	defer rNode.wal.Close()

	// INFO: 提交 proposeC 到 raft 里
	go func() {
		confChangeCount := uint64(0)
		for {
			select {
			case propose := <-rNode.proposeC:
				err := rNode.node.Propose(context.TODO(), []byte(propose))
				if err != nil {
					klog.Errorf(fmt.Sprintf("[raft proposeC]propose err:%v", err))
				}
				return

			case confChange := <-rNode.confChangeC:
				confChangeCount++
				confChange.ID = confChangeCount
				rNode.node.ProposeConfChange(context.TODO(), confChange)

			case <-rNode.stopc:
				rNode.stop()
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
			rNode.node.Tick()

			// store raft entries to wal, then publish over commit channel
		case rd := <-rNode.node.Ready():
			rNode.wal.Save(rd.HardState, rd.Entries)
			/*if !raft.IsEmptySnap(rd.Snapshot) {
				rNode.saveSnap(rd.Snapshot)
				rNode.raftStorage.ApplySnapshot(rd.Snapshot)
				rNode.publishSnapshot(rd.Snapshot)
			}*/
			rNode.raftStorage.Append(rd.Entries)
			rNode.transport.Send(rd.Messages)
			if _, ok := rNode.publishEntries(rNode.entriesToApply(rd.CommittedEntries)); !ok {
				rNode.stop()
				return
			}
			//rNode.maybeTriggerSnapshot(applyDoneC)
			rNode.node.Advance()

		case err := <-rNode.transport.ErrorC:
			klog.Errorf(fmt.Sprintf("[raft readyC]transport err:%v", err))
			return

		case <-rNode.stopc:
			rNode.stop()
			return
		}
	}
}

// TODO: 没理解？？？
func (rNode *raftNode) entriesToApply(ents []raftpb.Entry) (nents []raftpb.Entry) {
	if len(ents) == 0 {
		return ents
	}
	firstIdx := ents[0].Index
	if firstIdx > rNode.appliedIndex+1 {
		klog.Fatalf(fmt.Sprintf("first index of committed entry[%d] should <= progress.appliedIndex[%d]+1", firstIdx, rNode.appliedIndex))
	}
	if rNode.appliedIndex-firstIdx+1 < uint64(len(ents)) {
		nents = ents[rNode.appliedIndex-firstIdx+1:]
	}
	return nents
}

type commit struct {
	data       []string
	applyDoneC chan<- struct{}
}

// publishEntries writes committed log entries to commit channel and returns
// whether all entries could be published.
func (rNode *raftNode) publishEntries(ents []raftpb.Entry) (<-chan struct{}, bool) {
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
			//rNode.confState = *rNode.node.ApplyConfChange(cc)
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					rNode.transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(rNode.id) {
					klog.Infof("I've been removed from the cluster! Shutting down.")
					return nil, false
				}
				rNode.transport.RemovePeer(types.ID(cc.NodeID))
			}
		}
	}

	var applyDoneC chan struct{}

	if len(data) > 0 {
		applyDoneC = make(chan struct{}, 1)
		select {
		case rNode.commitC <- &commit{data, applyDoneC}:
		case <-rNode.stopc:
			return nil, false
		}
	}

	// after commit, update appliedIndex
	rNode.appliedIndex = ents[len(ents)-1].Index

	return applyDoneC, true
}

func (rNode *raftNode) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	klog.Infof("publishing snapshot at index %d", rNode.snapshotIndex)
	defer klog.Infof("finished publishing snapshot at index %d", rNode.snapshotIndex)

	if snapshotToSave.Metadata.Index <= rNode.appliedIndex {
		klog.Fatalf("snapshot index [%d] should > progress.appliedIndex [%d]", snapshotToSave.Metadata.Index, rNode.appliedIndex)
	}

	rNode.commitC <- nil // trigger kvstore to load snapshot

	rNode.confState = snapshotToSave.Metadata.ConfState
	rNode.snapshotIndex = snapshotToSave.Metadata.Index
	rNode.appliedIndex = snapshotToSave.Metadata.Index
}

/*func (rNode *raftNode) saveSnap(snap raftpb.Snapshot) error {
	// must save the snapshot index to the WAL before saving the
	// snapshot to maintain the invariant that we only Open the
	// wal at previously-saved snapshot indexes.
	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}
	if err := rNode.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}
	if err := rNode.snapshotter.SaveSnap(snap); err != nil {
		return err
	}

	return rNode.wal.ReleaseLockTo(snap.Metadata.Index)
}*/

// stop closes http, closes all channels, and stops raft.
func (rNode *raftNode) stop() {
	rNode.stopHTTP()
	close(rNode.commitC)
	close(rNode.errorC)
	rNode.node.Stop()
}

func (rNode *raftNode) stopHTTP() {

}

func (rNode *raftNode) loadSnapshot() *raftpb.Snapshot {
	snapshot, err := rNode.snapshotter.Load()
	if err != nil && err != snap.ErrNoSnapshot {
		klog.Fatalf("raftexample: error loading snapshot (%v)", err)
	}
	return snapshot
}

// openWAL returns a WAL ready for reading.
func (rNode *raftNode) openWAL(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(rNode.waldir) {
		if err := os.Mkdir(rNode.waldir, 0750); err != nil {
			klog.Fatalf("raftexample: cannot create dir for wal (%v)", err)
		}

		w, err := wal.Create(rNode.waldir, nil)
		if err != nil {
			klog.Fatalf("raftexample: create wal error (%v)", err)
		}
		w.Close()
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	klog.Infof("loading WAL at term %d and index %d", walsnap.Term, walsnap.Index)
	w, err := wal.Open(rNode.waldir, walsnap)
	if err != nil {
		klog.Fatalf("raftexample: error loading wal (%v)", err)
	}

	return w
}

// replayWAL replays WAL entries into the raft instance.
func (rNode *raftNode) replayWAL() *wal.WAL {
	klog.Infof("replaying WAL of member %d", rNode.id)
	snapshot := rNode.loadSnapshot()
	w := rNode.openWAL(snapshot)
	_, st, ents, err := w.ReadAll()
	if err != nil {
		klog.Fatalf("raftexample: failed to read WAL (%v)", err)
	}
	rNode.raftStorage = raft.NewMemoryStorage()
	/*if snapshot != nil {
		rNode.raftStorage.ApplySnapshot(*snapshot)
	}*/
	rNode.raftStorage.SetHardState(st)

	// append to storage so raft starts at the right place in log
	rNode.raftStorage.Append(ents)
	// send nil once lastIndex is published so client knows commit channel is current
	if len(ents) > 0 {
		rNode.lastIndex = ents[len(ents)-1].Index
	} else {
		rNode.commitC <- nil
	}
	return w
}

func (rNode *raftNode) Process(ctx context.Context, m raftpb.Message) error {
	return rNode.node.Step(ctx, m)
}
func (rNode *raftNode) IsIDRemoved(id uint64) bool                               { return false }
func (rNode *raftNode) ReportUnreachable(id uint64)                              {}
func (rNode *raftNode) ReportSnapshot(id uint64, status etcdRaft.SnapshotStatus) {}
