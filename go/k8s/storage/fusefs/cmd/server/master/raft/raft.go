package raft

import (
	"go.uber.org/zap"
	"k8s.io/klog/v2"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.etcd.io/etcd/server/v3/wal"
)

type Config struct {
	heartbeat time.Duration // 1000ms
}

type RaftNode struct {
	tickMu *sync.Mutex

	ticker *time.Ticker

	raft.Node
	wal       *wal.WAL
	transport *rafthttp.Transport
}

func NewRaftNode(config *Config) *RaftNode {
	r := &RaftNode{
		Node:   nil,
		ticker: time.NewTicker(config.heartbeat),
	}

	return r
}

func (r *RaftNode) start() {

	oldwal := wal.Exist(r.waldir)
	r.wal = r.replayWAL()

	rpeers := make([]raft.Peer, len(r.peers))
	for i := range rpeers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
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
	if oldwal {
		r.Node = raft.RestartNode(c)
	} else {
		r.Node = raft.StartNode(c, rpeers)
	}

	r.transport = &rafthttp.Transport{
		Logger:      r.logger,
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

	go r.serveRaft()
	go r.serveChannels()
}

func (r *RaftNode) tick() {
	r.tickMu.Lock()
	r.Tick()
	r.tickMu.Unlock()
}

func (r *RaftNode) serveRaft() {
	url, err := url.Parse(r.peers[r.id-1])
	if err != nil {
		klog.Fatalf("raftexample: Failed parsing URL (%v)", err)
	}

	ln, err := newStoppableListener(url.Host, r.httpstopc)
	if err != nil {
		klog.Fatalf("raftexample: Failed to listen rafthttp (%v)", err)
	}

	err = (&http.Server{Handler: r.transport.Handler()}).Serve(ln)
	select {
	case <-r.httpstopc:
	default:
		klog.Fatalf("raftexample: Failed to serve rafthttp (%v)", err)
	}
	close(r.httpdonec)
}

func (r *RaftNode) serveChannels() {

	// event loop on raft state machine updates
	for {
		select {
		case <-r.ticker.C:
			r.tick()

		// store raft entries to wal, then publish over commit channel
		case rd := <-r.Ready():
			r.wal.Save(rd.HardState, rd.Entries)
			if !raft.IsEmptySnap(rd.Snapshot) {
				r.saveSnap(rd.Snapshot)
				r.raftStorage.ApplySnapshot(rd.Snapshot)
				r.publishSnapshot(rd.Snapshot)
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
			r.writeError(err)
			return

		case <-r.stopc:
			r.stop()
			return
		}
	}
}
