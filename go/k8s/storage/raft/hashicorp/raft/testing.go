package raft

import (
	"bytes"
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// LoopbackTransport is an interface that provides a loopback transport suitable for testing
// It's there so we don't have to rewrite tests.
type LoopbackTransport interface {
	Transport // Embedded transport reference
	WithPeers // Embedded peer management
	WithClose // with a close routine
}

// WithPeers is an interface that a transport may provide which allows for connection and
// disconnection. Unless the transport is a loopback transport, the transport specified to
// "Connect" is likely to be nil.
type WithPeers interface {
	Connect(peer ServerAddress, t Transport) // Connect a peer
	Disconnect(peer ServerAddress)           // Disconnect a given peer
	DisconnectAll()                          // Disconnect all peers, possibly to reconnect them later
}

type ClusterConfig struct {
	Conf      *Config
	Peers     []string
	Bootstrap bool
}

type Cluster struct {
	conf *Config

	raftDirs   []string
	rafts      []*Raft
	stores     []*MemoryStore
	fsms       []FSM
	snapshots  []*FileSnapshotStore
	transports []LoopbackTransport

	longstopTimeout time.Duration

	observationCh chan Observation
}

func NewCluster(config *ClusterConfig) *Cluster {
	if config.Conf == nil {
		config.Conf = DefaultConfig()
	}

	cluster := &Cluster{
		conf:            config.Conf,
		longstopTimeout: 5 * time.Second,
		observationCh:   make(chan Observation, 1024),
	}

	var configuration Configuration
	for i := 0; i < len(config.Peers); i++ {
		peerInfo := config.Peers[i]
		peer := strings.Split(peerInfo, "/")
		localID := ServerID(peer[0])
		configuration.Servers = append(configuration.Servers, Server{
			Suffrage: Voter,
			ID:       localID,
			Address:  ServerAddress(peer[1]),
		})

		transport := NewMemoryTransport(ServerAddress(peer[1]))
		cluster.transports = append(cluster.transports, transport)
	}

	// Wire the transports together
	cluster.FullyConnect()

	// Create all the rafts
	for i := 0; i < len(config.Peers); i++ {
		raftDir := getRaftDir(string(configuration.Servers[i].ID))
		cluster.raftDirs = append(cluster.raftDirs, raftDir)

		snapStore, err := NewFileSnapshotStore(raftDir, 5)
		if err != nil {
			klog.Fatalf(fmt.Sprintf("NewFileSnapshotStore failed: %v", err))
		}
		store := NewMemoryStore()
		peerConf := config.Conf
		peerConf.LocalID = configuration.Servers[i].ID
		if config.Bootstrap {
			err := BootstrapCluster(peerConf, store, store, snapStore, configuration)
			if err != nil {
				klog.Fatalf(fmt.Sprintf("BootstrapCluster failed: %v", err))
			}
		}
		fsm := &MockFSM{}
		cluster.fsms = append(cluster.fsms, fsm)
		raft, err := NewRaft(peerConf, fsm, store, store, snapStore, cluster.transports[i])
		if err != nil {
			klog.Fatalf(fmt.Sprintf("NewRaft failed: %v", err))
		}

		raft.RegisterObserver(NewObserver(cluster.observationCh, false, nil))
		if err != nil {
			klog.Fatalf(fmt.Sprintf("RegisterObserver failed: %v", err))
		}
		cluster.rafts = append(cluster.rafts, raft)
	}

	return cluster
}

// EnsureLeader checks that ALL the nodes think the leader is the given expected leader
func (cluster *Cluster) EnsureLeader(expect ServerAddress) {
	if len(expect) == 0 {
		klog.Fatal("no leader")
	}

	fail := false
	for _, r := range cluster.rafts {
		leader := r.Leader()
		if leader != expect {
			if len(leader) == 0 {
				leader = "[none]"
			}
			klog.Errorf(fmt.Sprintf("expected leader:%s got:%s", expect, leader))
			fail = true
		}
	}
	if fail {
		klog.Fatalf("at least one peer has the wrong notion of leader")
	}

	klog.Infof(fmt.Sprintf("expected %s is leader", string(expect)))
}

func (cluster *Cluster) EnsureFSMSame() {
	limit := time.Now().Add(cluster.longstopTimeout)
	first := getMockFSM(cluster.fsms[0])

CHECK:
	first.Lock()
	for i, fsmRaw := range cluster.fsms {
		fsm := getMockFSM(fsmRaw)
		if i == 0 {
			continue
		}
		fsm.Lock()

		if len(first.logs) != len(fsm.logs) {
			fsm.Unlock()
			if time.Now().After(limit) {
				klog.Fatalf(fmt.Sprintf("FSM log length mismatch: %d %d",
					len(first.logs), len(fsm.logs)))
			} else {
				goto WAIT
			}
		}

		for idx := 0; idx < len(first.logs); idx++ {
			if bytes.Compare(first.logs[idx], fsm.logs[idx]) != 0 {
				fsm.Unlock()
				if time.Now().After(limit) {
					klog.Fatalf(fmt.Sprintf("FSM log mismatch at index %d", idx))
				} else {
					goto WAIT
				}
			}
		}
		if len(first.configurations) != len(fsm.configurations) {
			fsm.Unlock()
			if time.Now().After(limit) {
				klog.Fatalf(fmt.Sprintf("FSM configuration length mismatch: %d %d",
					len(first.logs), len(fsm.logs)))
			} else {
				goto WAIT
			}
		}

		for idx := 0; idx < len(first.configurations); idx++ {
			if !reflect.DeepEqual(first.configurations[idx], fsm.configurations[idx]) {
				fsm.Unlock()
				if time.Now().After(limit) {
					klog.Fatalf(fmt.Sprintf("FSM configuration mismatch at index %d: %v, %v",
						idx, first.configurations[idx], fsm.configurations[idx]))
				} else {
					goto WAIT
				}
			}
		}
		fsm.Unlock()
	}

	first.Unlock()
	for _, log := range first.logs {
		klog.Infof(fmt.Sprintf("log in fsm is %s", string(log)))
	}
	return

WAIT:
	first.Unlock()
	cluster.WaitEvent(nil, cluster.conf.CommitTimeout)
	goto CHECK
}

// getConfiguration returns the configuration of the given Raft instance, or
// fails the test if there's an error
func (cluster *Cluster) getConfiguration(r *Raft) Configuration {
	future := r.GetConfiguration()
	if err := future.Error(); err != nil {
		klog.Fatalf(fmt.Sprintf("failed to get configuration: %v", err))
		return Configuration{}
	}

	return future.Configuration()
}

// EnsurePeersSame makes sure all the rafts have the same set of peers.
func (cluster *Cluster) EnsurePeersSame() {
	limit := time.Now().Add(cluster.longstopTimeout)
	peerSet := cluster.getConfiguration(cluster.rafts[0])

CHECK:
	for i, raft := range cluster.rafts {
		if i == 0 {
			continue
		}

		otherSet := cluster.getConfiguration(raft)
		if !reflect.DeepEqual(peerSet, otherSet) {
			if time.Now().After(limit) {
				klog.Fatalf(fmt.Sprintf("peer mismatch: %+v %+v", peerSet, otherSet))
			} else {
				goto WAIT
			}
		}
	}
	return

WAIT:
	cluster.WaitEvent(nil, cluster.conf.CommitTimeout)
	goto CHECK
}

// Leader waits for the cluster to elect a leader and stay in a stable state.
func (cluster *Cluster) Leader() *Raft {
	leaders := cluster.GetInState(Leader)
	if len(leaders) != 1 {
		klog.Fatalf(fmt.Sprintf("expected one leader: %v", leaders))
	}
	return leaders[0]
}

// Followers waits for the cluster to have N-1 followers and stay in a stable
// state.
func (cluster *Cluster) Followers() []*Raft {
	expFollowers := len(cluster.rafts) - 1
	followers := cluster.GetInState(Follower)
	if len(followers) != expFollowers {
		klog.Fatalf(fmt.Sprintf("timeout waiting for %d followers (followers are %v)", expFollowers, followers))
	}
	return followers
}

// GetInState polls the state of the cluster and attempts to identify when it has
// settled into the given state.
func (cluster *Cluster) GetInState(s RaftState) []*Raft {
	timeout := cluster.conf.HeartbeatTimeout
	if timeout < cluster.conf.ElectionTimeout {
		timeout = cluster.conf.ElectionTimeout
	}
	timeout = 2*timeout + cluster.conf.CommitTimeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		inState, highestTerm := cluster.pollState(s)
		if highestTerm == 0 {
			timer.Reset(cluster.longstopTimeout)
		} else {
			timer.Reset(timeout)
		}

		select {
		case <-timer.C:
			return inState
		}
	}
}

func (cluster *Cluster) pollState(s RaftState) ([]*Raft, uint64) {
	var highestTerm uint64
	in := make([]*Raft, 0, 1)
	for _, r := range cluster.rafts {
		if r.State() == s {
			in = append(in, r)
		}
		term := r.getCurrentTerm()
		if term > highestTerm {
			highestTerm = term
		}
	}
	return in, highestTerm
}

func (cluster *Cluster) FullyConnect() {
	klog.Infof("fully connecting")
	for i, t1 := range cluster.transports {
		for j, t2 := range cluster.transports {
			if i != j {
				t1.Connect(t2.LocalAddr(), t2)
				t2.Connect(t1.LocalAddr(), t1)
			}
		}
	}
}

// WaitForReplication blocks until every FSM in the cluster has the given
// length, or the long sanity check timeout expires.
func (cluster *Cluster) WaitForReplication(fsmLength int) {
	limitCh := time.After(cluster.longstopTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), cluster.conf.CommitTimeout)
	defer cancel()

CHECK:
	for {
		select {
		case <-limitCh:
			klog.Fatalf("timeout waiting for replication")

		case <-cluster.WaitEventChan(ctx, nil):
			for _, fsmRaw := range cluster.fsms {
				fsm := getMockFSM(fsmRaw)
				fsm.Lock()
				num := len(fsm.logs)
				fsm.Unlock()
				if num != fsmLength {
					continue CHECK
				}
			}
			return
		}
	}
}

func getMockFSM(fsm FSM) *MockFSM {
	switch f := fsm.(type) {
	case *MockFSM:
		return f
	default:
		return nil
	}
}

// WaitEvent waits until an observation is made, a timeout occurs, or a test
// failure is signaled. It is possible to set a filter to look for specific
// observations. Setting timeout to 0 means that it will wait forever until a
// non-filtered observation is made or a test failure is signaled.
func (cluster *Cluster) WaitEvent(filter FilterFn, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	eventCh := cluster.WaitEventChan(ctx, filter)
	select {
	case <-eventCh:
	}
}

func (cluster *Cluster) WaitEventChan(ctx context.Context, filter FilterFn) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case o, ok := <-cluster.observationCh:
				if !ok || filter == nil || filter(&o) {
					return
				}
			}
		}
	}()
	return ch
}

// Close shuts down the cluster and cleans up.
func (cluster *Cluster) Close() {
	var futures []Future
	for _, raft := range cluster.rafts {
		futures = append(futures, raft.Shutdown())
	}

	for _, dir := range cluster.raftDirs {
		os.RemoveAll(dir)
	}

	for _, f := range futures {
		if err := f.Error(); err != nil {
			klog.Fatalf(fmt.Sprintf("shutdown future err: %v", err))
		}
	}
}

func getRaftDir(raftId string) string {
	raftDir := fmt.Sprintf("raft/raft_%s", raftId)
	raftDir, _ = filepath.Abs(raftDir)
	if _, err := os.Stat(raftDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(raftDir, 0777)
		} else {
			klog.Fatalf(fmt.Sprintf("%s is err:%v", raftDir, err))
		}
	}

	return raftDir
}
