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
}
