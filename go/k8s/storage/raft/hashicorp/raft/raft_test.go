package raft

import (
	"fmt"
	"k8s.io/klog/v2"
	"testing"
	"time"
)

func TestRaftStartStop(test *testing.T) {
	cluster := NewCluster(&ClusterConfig{
		Peers: []string{
			"1/127.0.0.1:7000",
			"2/127.0.0.1:8000",
			"3/127.0.0.1:9000",
		},
		Bootstrap: true,
	})

	// wait for leader election
	time.Sleep(time.Second * 10)
	cluster.Close()

	raft := cluster.rafts[0]
	// Everything should fail now
	if f := raft.Apply(nil, 0); f.Error() != ErrRaftShutdown {
		test.Fatalf("should be shutdown: %v", f.Error())
	}
	// Should be idempotent
	if f := raft.Shutdown(); f.Error() != nil {
		test.Fatalf("shutdown should be idempotent")
	}
}

func TestRaftLiveBootstrap(test *testing.T) {
	cluster := NewCluster(&ClusterConfig{
		Peers: []string{
			"1/127.0.0.1:7000",
			"2/127.0.0.1:8000",
			"3/127.0.0.1:9000",
		},
		Bootstrap: false,
	})
	defer cluster.Close()

	// Bootstrap one of the nodes live.
	configuration := Configuration{}
	for _, r := range cluster.rafts {
		server := Server{
			ID:      r.localID,
			Address: r.localAddr,
		}
		configuration.Servers = append(configuration.Servers, server)
	}
	boot := cluster.rafts[0].BootstrapCluster(configuration)
	if err := boot.Error(); err != nil {
		test.Fatalf("bootstrap err: %v", err)
	}

	// leader election finished
	time.Sleep(time.Second * 5)

	// Should be one leader.
	cluster.Followers()
	leader := cluster.Leader()
	cluster.EnsureLeader(leader.localAddr)

	// Should be able to apply.
	future := leader.Apply([]byte("test"), cluster.conf.CommitTimeout)
	if err := future.Error(); err != nil {
		klog.Fatalf(fmt.Sprintf("apply err: %v", err))
	}
	cluster.WaitForReplication(1)

	// Make sure the live bootstrap fails now that things are started up.
	boot = cluster.rafts[0].BootstrapCluster(configuration)
	if err := boot.Error(); err != ErrCantBootstrap {
		klog.Fatalf(fmt.Sprintf("bootstrap should have failed: %v", err))
	}
}
