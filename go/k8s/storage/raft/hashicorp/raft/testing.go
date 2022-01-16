package raft

import (
	"fmt"
	"k8s.io/klog/v2"
)

type ClusterConfig struct {
	Conf      *Config
	Peers     int
	Bootstrap bool
}

type Cluster struct {
	rafts      []*Raft
	stores     []*MemoryStore
	fsms       []FSM
	snapshots  []*FileSnapshotStore
	transports []LoopbackTransport

	observationCh chan Observation
}

func MakeCluster(config *ClusterConfig) *Cluster {
	if config.Conf == nil {
		config.Conf = DefaultConfig()
	}

	cluster := &Cluster{
		observationCh: make(chan Observation, 1024),
	}

	var configuration Configuration
	// Setup the stores and transports
	for i := 0; i < config.Peers; i++ {

		localID := ServerID(fmt.Sprintf("server-%s", addr))
		configuration.Servers = append(configuration.Servers, Server{
			Suffrage: Voter,
			ID:       localID,
			Address:  addr,
		})
	}

	// Create all the rafts
	for i := 0; i < config.Peers; i++ {
		logs := cluster.stores[i]
		store := cluster.stores[i]
		snap := cluster.snapshots[i]
		trans := cluster.transports[i]
		peerConf := config.Conf
		peerConf.LocalID = configuration.Servers[i].ID
		if config.Bootstrap {
			err := BootstrapCluster(peerConf, logs, store, snap, configuration)
			if err != nil {
				klog.Fatalf(fmt.Sprintf("BootstrapCluster failed: %v", err))
			}
		}

		raft, err := NewRaft(peerConf, cluster.fsms[i], logs, store, snap, trans)
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

// Close shuts down the cluster and cleans up.
func (cluster *Cluster) Close() {
	var futures []Future
	for _, r := range cluster.rafts {
		futures = append(futures, r.Shutdown())
	}

}
