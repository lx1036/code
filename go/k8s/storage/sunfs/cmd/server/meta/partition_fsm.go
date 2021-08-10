package meta

// INFO: 是raft statemachine 实现 https://github.com/tiglabs/raft/blob/master/statemachine.go#L22-L30

import (
	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
)

func (partition *metaPartition) Apply(command []byte, index uint64) (interface{}, error) {
	panic("implement me")
}

func (partition *metaPartition) ApplyMemberChange(confChange *proto.ConfChange, index uint64) (interface{}, error) {
	panic("implement me")
}

func (partition *metaPartition) Snapshot() (proto.Snapshot, error) {
	panic("implement me")
}

func (partition *metaPartition) ApplySnapshot(peers []proto.Peer, iter proto.SnapIterator) error {
	panic("implement me")
}

func (partition *metaPartition) HandleFatalEvent(err *raft.FatalError) {

}

func (partition *metaPartition) HandleLeaderChange(leader uint64) {

}
