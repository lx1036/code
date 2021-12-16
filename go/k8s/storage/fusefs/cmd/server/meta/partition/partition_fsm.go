package partition

// INFO: 是raft statemachine 实现 https://github.com/tiglabs/raft/blob/master/statemachine.go#L22-L30

import (
	"encoding/json"
	"fmt"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/tiglabs/raft"
	raftproto "github.com/tiglabs/raft/proto"
	"k8s.io/klog/v2"
)

const (
	opFSMCreateInode uint32 = iota
	opFSMUnlinkInode
	opFSMCreateDentry
	opFSMDeleteDentry
	opFSMDeletePartition
	opFSMUpdatePartition
	opFSMDecommissionPartition
	opFSMStoreTick
	startStoreTick
	stopStoreTick
	opFSMUpdateDentry
	opFSMCreateLinkInode
	opFSMEvictInode
	opFSMSetAttr
)

func (partition *MetaPartitionFSM) Apply(command []byte, index uint64) (resp interface{}, err error) {
	msg := &MetaItem{}
	defer func() {
		if err == nil {
			partition.saveApplyID(index)
		}
	}()
	if err = json.Unmarshal(command, msg); err != nil {
		return
	}

	switch msg.Op {
	case opFSMCreateInode:
		ino := NewInode(0, 0)
		if err = ino.Unmarshal(msg.V); err != nil {
			return
		}
		if partition.config.Cursor < ino.Inode {
			partition.config.Cursor = ino.Inode
		}
		partition.inodeTree.tree.ReplaceOrInsert(ino)

	case opFSMEvictInode:

	case opFSMCreateDentry:
	case opFSMDeleteDentry:
	case opFSMUpdateDentry:

	case opFSMCreateLinkInode:
	case opFSMUnlinkInode:

	case opFSMSetAttr:

	case opFSMUpdatePartition:

	case opFSMStoreTick: // 用来 snapshot，把 inode/dentry btree 持久化到文件
		inodeTree := partition.getInodeTree()
		dentryTree := partition.getDentryTree()
		msg := &storeMsg{
			command:    opFSMStoreTick,
			applyIndex: index,
			inodeTree:  inodeTree,
			dentryTree: dentryTree,
		}

		partition.storeChan <- msg
	}

	klog.Infof(fmt.Sprintf("[apply raft sm]%+v", *msg))
	return
}

func (partition *MetaPartitionFSM) ApplyMemberChange(confChange *raftproto.ConfChange, index uint64) (interface{}, error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) Snapshot() (raftproto.Snapshot, error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) ApplySnapshot(peers []raftproto.Peer, iter raftproto.SnapIterator) error {
	panic("implement me")
}

func (partition *MetaPartitionFSM) HandleFatalEvent(err *raft.FatalError) {

}

func (partition *MetaPartitionFSM) HandleLeaderChange(leader uint64) {

}

// Put INFO: 提交 key-value 到 状态机，@see MetaPartitionFSM.Apply()
func (partition *MetaPartitionFSM) Put(key, value interface{}) (resp interface{}, err error) {
	entry := NewMetaItem(opFSMCreateInode, nil, nil)
	entry.Op = key.(uint32)
	if value != nil {
		entry.V = value.([]byte)
	}
	cmd, err := json.Marshal(entry)
	if err != nil {
		return
	}

	// submit to the raft store
	resp, err = partition.raftPartition.Submit(cmd)
	return
}

func (partition *MetaPartitionFSM) Get(key interface{}) (interface{}, error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) Del(key interface{}) (interface{}, error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) saveApplyID(applyId uint64) {
	atomic.StoreUint64(&partition.applyID, applyId)
}

func (partition *MetaPartitionFSM) ChangeMember(changeType raftproto.ConfChangeType, peer raftproto.Peer, context []byte) (resp interface{}, err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) IsLeader() (leaderAddr string, isLeader bool) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) LoadSnapshotSign(p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) DeleteRaft() error {
	panic("implement me")
}

func (partition *MetaPartitionFSM) IsExistPeer(peer raftproto.Peer) bool {
	panic("implement me")
}

func (partition *MetaPartitionFSM) TryToLeader(groupID uint64) error {
	panic("implement me")
}

func (partition *MetaPartitionFSM) CanRemoveRaftMember(peer raftproto.Peer) error {
	panic("implement me")
}

func (partition *MetaPartitionFSM) getRaftPort() (heartbeat, replica int, err error) {
	raftConfig := partition.config.RaftStore.RaftConfig()
	heartbeatAddrSplits := strings.Split(raftConfig.HeartbeatAddr, ":")
	replicaAddrSplits := strings.Split(raftConfig.ReplicateAddr, ":")
	if len(heartbeatAddrSplits) != 2 {
		err = ErrIllegalHeartbeatAddress
		return
	}
	if len(replicaAddrSplits) != 2 {
		err = ErrIllegalReplicateAddress
		return
	}
	heartbeat, err = strconv.Atoi(heartbeatAddrSplits[1])
	if err != nil {
		return
	}
	replica, err = strconv.Atoi(replicaAddrSplits[1])
	if err != nil {
		return
	}
	return
}
