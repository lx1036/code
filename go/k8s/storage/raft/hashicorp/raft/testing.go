package raft

import (
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strings"
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
	raftDirs   []string
	rafts      []*Raft
	stores     []*MemoryStore
	fsms       []FSM
	snapshots  []*FileSnapshotStore
	transports []LoopbackTransport

	observationCh chan Observation
}

func NewCluster(config *ClusterConfig) *Cluster {
	if config.Conf == nil {
		config.Conf = DefaultConfig()
	}

	cluster := &Cluster{
		observationCh: make(chan Observation, 1024),
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
