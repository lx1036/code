package pkg

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/pkg/fileutil"
	"go.uber.org/zap"
	"k8s.io/klog/v2"
	"os"
	"time"

	"go.etcd.io/etcd/etcdserver/api/rafthttp"
	"go.etcd.io/etcd/etcdserver/api/snap"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"go.etcd.io/etcd/wal"
)

const (
	defaultSnapshotCount uint64 = 10000
)

// A key-value stream backed by raft
type raftNode struct {
	proposeC    <-chan string            // proposed messages (k,v)
	confChangeC <-chan raftpb.ConfChange // proposed cluster config changes
	commitC     chan<- *string           // entries committed to log (k,v)
	errorC      chan<- error             // errors from raft session

	id          int      // client ID for raft session
	peers       []string // raft peer URLs
	join        bool     // node is joining an existing cluster
	waldir      string   // path to WAL directory
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

	snapshotter      *snap.Snapshotter
	snapshotterReady chan *snap.Snapshotter // signals when snapshotter is ready

	snapCount uint64
	transport *rafthttp.Transport
	stopc     chan struct{} // signals proposal channel closed
	httpstopc chan struct{} // signals http server to shutdown
	httpdonec chan struct{} // signals http server shutdown complete
}

func (rNode *raftNode) serveChannels() {

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// send proposals over raft
	go func() {
		confChangeCount := uint64(0)
		for rNode.proposeC != nil && rNode.confChangeC != nil {
			select {
			case propose, ok := <-rNode.proposeC:
				if !ok {
					rNode.proposeC = nil
				} else {
					// blocks until accepted by raft state machine
					// proposes key-value data be appended to the log.
					err := rNode.node.Propose(context.TODO(), []byte(propose))
					if err != nil {
						klog.Error(err)
					}
				}
			case confChange, ok := <-rNode.confChangeC:
				if !ok {
					rNode.confChangeC = nil
				} else {
					// ??? confChange
					confChangeCount++
					confChange.ID = confChangeCount
					err := rNode.node.ProposeConfChange(context.TODO(), confChange)
					if err != nil {
						klog.Error(err)
					}
				}
			}
		}

		// client closed channel; shutdown raft if not already
		close(rNode.stopc)
	}()

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			rNode.node.Tick()

			// store raft entries to wal, then publish over commit channel
		case rd := <-rNode.node.Ready():

		case <-rNode.stopc:
			rNode.stop()
			return
		}
	}
}

// stop closes http, closes all channels, and stops raft.
func (rNode *raftNode) stop() {
	rNode.stopHTTP()
	close(rNode.commitC)
	close(rNode.errorC)
	rNode.node.Stop()
}

func (rNode *raftNode) stopHTTP() {

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

	go rNode.serveRaft()
	go rNode.serveChannels()
}

// newRaftNode initiates a raft instance and returns a committed log entry
// channel and error channel. Proposals for log updates are sent over the
// provided the proposal channel. All log entries are replayed over the
// commit channel, followed by a nil message (to indicate the channel is
// current), then new log entries. To shutdown, close proposeC and read errorC.
func NewRaftNode(id int, peers []string, join bool, getSnapshot func() ([]byte, error), proposeC <-chan string,
	confChangeC <-chan raftpb.ConfChange) (<-chan *string, <-chan error, <-chan *snap.Snapshotter) {
	commitC := make(chan *string)
	errorC := make(chan error)

	node := &raftNode{
		proposeC:    proposeC,
		confChangeC: confChangeC,
		commitC:     commitC,
		errorC:      errorC,
		id:          id,
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
