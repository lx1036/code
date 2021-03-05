package raftstore

import "github.com/tiglabs/raft"

// NodeManager defines the necessary methods for node address management.
type NodeManager interface {
	// add node address with specified port.
	AddNodeWithPort(nodeID uint64, addr string, heartbeat int, replicate int)

	// delete node address information
	DeleteNode(nodeID uint64)
}

// NodeResolver defines the methods for node address resolving and management.
// It is extended from SocketResolver and NodeManager.
type NodeResolver interface {
	raft.SocketResolver
	NodeManager
}
