package meta

// INFO: 是raft statemachine 实现 https://github.com/tiglabs/raft/blob/master/statemachine.go#L22-L30

import (
	"encoding/json"
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

// Put puts the given key-value pair (operation key and operation request) into the raft store
func (partition *metaPartition) Put(key, value interface{}) (interface{}, error) {

	entry := NewMetaItem(0, nil, nil)
	entry.Op = key.(uint32)
	if value != nil {
		entry.V = value.([]byte)
	}
	cmd, err := json.Marshal(entry)
	if err != nil {
		return
	}

	// submit to the raft store
	resp, err := partition.raftPartition.Submit(cmd)

	return
}

func (partition *metaPartition) Get(key interface{}) (interface{}, error) {
	panic("implement me")
}

func (partition *metaPartition) Del(key interface{}) (interface{}, error) {
	panic("implement me")
}
