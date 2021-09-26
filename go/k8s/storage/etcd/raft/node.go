package raft

import (
	"context"
	pb "go.etcd.io/etcd/raft/v3/raftpb"
)

var (
	emptyState = pb.HardState{}
)

type Node interface {

	// Campaign INFO: Follower 变成 Candidate, 竞选成 Leader
	Campaign(ctx context.Context) error
}

// INFO: node in raft cluster
type node struct {
}

func IsEmptyHardState(st pb.HardState) bool {
	return isHardStateEqual(st, emptyState)
}

func isHardStateEqual(a, b pb.HardState) bool {
	return a.Term == b.Term && a.Vote == b.Vote && a.Commit == b.Commit
}
