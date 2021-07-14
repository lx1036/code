package raft

import (
	"k8s-lx1036/k8s/storage/sunfs/pkg/raft/proto"
)

// Transport raft server transport
type Transport interface {
	Send(m *proto.Message)
	SendSnapshot(m *proto.Message, rs *snapshotStatus)
	Stop()
}

// The SocketResolver interface is supplied by the application to resolve NodeID to net.Addr addresses.
type SocketResolver interface {
	NodeAddress(nodeID uint64, stype SocketType) (addr string, err error)
}

type SocketType byte

const (
	HeartBeat SocketType = 0
	Replicate SocketType = 1
)

func (t SocketType) String() string {
	switch t {
	case 0:
		return "HeartBeat"
	case 1:
		return "Replicate"
	}
	return "unkown"
}
