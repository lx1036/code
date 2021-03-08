package raftstore

import (
	"github.com/tiglabs/raft"
	"sync"
)

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

// Default thread-safe implementation of the NodeResolver interface.
type nodeResolver struct {
	nodeMap sync.Map
}

func (resolver *nodeResolver) NodeAddress(nodeID uint64, stype raft.SocketType) (addr string, err error) {
	panic("implement me")
}

func (resolver *nodeResolver) AddNodeWithPort(nodeID uint64, addr string, heartbeat int, replicate int) {
	panic("implement me")
}

func (resolver *nodeResolver) DeleteNode(nodeID uint64) {
	panic("implement me")
}

// NewNodeResolver returns a new NodeResolver instance for node address management and resolving.
func NewNodeResolver() NodeResolver {
	return &nodeResolver{}
}
