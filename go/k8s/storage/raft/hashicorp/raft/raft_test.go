package raft

import (
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
