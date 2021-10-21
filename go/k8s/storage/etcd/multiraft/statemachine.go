package multiraft

import (
	"k8s-lx1036/k8s/storage/raft/proto"
)

// The StateMachine interface is supplied by the application to persist/snapshot data of application.
type StateMachine interface {
	Apply(command []byte, index uint64) (interface{}, error)
	ApplyMemberChange(confChange *proto.ConfChange, index uint64) (interface{}, error)
	Snapshot() (proto.Snapshot, error)
	ApplySnapshot(peers []proto.Peer, iter proto.SnapIterator) error
	//HandleFatalEvent(err *FatalError)
	HandleLeaderChange(leader uint64)
}
