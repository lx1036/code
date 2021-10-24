package raftstore

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"k8s-lx1036/k8s/storage/etcd/multiraft"
)

var (
	ErrNoSuchNode        = errors.New("no such node")
	ErrIllegalAddress    = errors.New("illegal address")
	ErrUnknownSocketType = errors.New("unknown socket type")
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
	multiraft.SocketResolver
	NodeManager
}

// This private struct defines the necessary properties for node address info.
type nodeAddress struct {
	Heartbeat string
	Replicate string
}

// Default thread-safe implementation of the NodeResolver interface.
type nodeResolver struct {
	nodeMap sync.Map
}

func (resolver *nodeResolver) NodeAddress(nodeID uint64, socketType multiraft.SocketType) (addr string, err error) {
	val, ok := resolver.nodeMap.Load(nodeID)
	if !ok {
		err = ErrNoSuchNode
		return
	}

	address, ok := val.(*nodeAddress)
	if !ok {
		err = ErrIllegalAddress
		return
	}
	switch socketType {
	case multiraft.HeartBeat:
		addr = address.Heartbeat
	case multiraft.Replicate:
		addr = address.Replicate
	default:
		err = ErrUnknownSocketType
	}

	return
}

func (resolver *nodeResolver) AddNodeWithPort(nodeID uint64, addr string, heartbeat int, replicate int) {
	if heartbeat == 0 {
		heartbeat = DefaultHeartbeatPort
	}
	if replicate == 0 {
		replicate = DefaultReplicaPort
	}
	if len(strings.TrimSpace(addr)) != 0 {
		resolver.nodeMap.Store(nodeID, &nodeAddress{
			Heartbeat: fmt.Sprintf("%s:%d", addr, heartbeat),
			Replicate: fmt.Sprintf("%s:%d", addr, replicate),
		})
	}
}

func (resolver *nodeResolver) DeleteNode(nodeID uint64) {
	resolver.nodeMap.Delete(nodeID)
}

// NewNodeResolver returns a new NodeResolver instance for node address management and resolving.
func NewNodeResolver() NodeResolver {
	return &nodeResolver{}
}
