package server

import (
	"k8s-lx1036/k8s/storage/etcd/raft"
)

// RaftNode INFO: 对 raft 包的封装，给 Server 调用
type RaftNode struct {
	raft.Node

	// a chan to send out readState
	readStateChan chan raft.ReadState
}

func newRaftNode(config *raft.Config, peers []raft.Peer) *RaftNode {
	node := raft.StartNode(config, peers)
	raftNode := &RaftNode{
		Node:          node,
		readStateChan: make(chan raft.ReadState, 1),
	}

	return raftNode
}

func (raftNode *RaftNode) start() {
	go func() {

		for {
			select {

			case <-raftNode.stopped:
				return
			}
		}

	}()
}
