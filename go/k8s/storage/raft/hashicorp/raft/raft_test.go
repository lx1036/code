package raft

import (
	"testing"
)

func TestRaftStartStop(test *testing.T) {
	cluster := MakeCluster(1, test, nil)
	cluster.Close()
}
