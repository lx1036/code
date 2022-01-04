package meta

// INFO: 是raft statemachine 实现 https://github.com/tiglabs/raft/blob/master/statemachine.go#L22-L30

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"github.com/tiglabs/raft"
	raftproto "github.com/tiglabs/raft/proto"
	"k8s.io/klog/v2"
)

const (
	opFSMCreateInode uint32 = iota
	opFSMUnlinkInode
	opFSMUnlinkInodeBatch
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

// INFO: 创建 raft partition，实际上每一个 partition 都有其自己的 wal path
func (partition *PartitionFSM) startRaft() error {
	var (
		err   error
		peers []raftstore.PeerAddress
	)
	raftConfig := partition.config.RaftStore.RaftConfig()
	heartbeatPort, _ := strconv.Atoi(strings.Split(raftConfig.HeartbeatAddr, ":")[1])
	replicaPort, _ := strconv.Atoi(strings.Split(raftConfig.ReplicateAddr, ":")[1])
	for _, peer := range partition.config.Peers {
		values := strings.Split(peer.Addr, ":") // 127.0.0.1:8500
		port, _ := strconv.Atoi(values[1])
		peers = append(peers, raftstore.PeerAddress{
			Peer: raftproto.Peer{
				ID: peer.ID,
			},
			Address:       values[0],
			Port:          port,
			HeartbeatPort: heartbeatPort,
			ReplicaPort:   replicaPort,
		})
	}
	partition.raftPartition, err = partition.config.RaftStore.CreatePartition(&raftstore.PartitionConfig{
		ID:      partition.config.PartitionId,
		Applied: partition.applyID,
		Peers:   peers,
		SM:      partition,
	})

	return err
}

func (partition *PartitionFSM) Apply(command []byte, index uint64) (resp interface{}, err error) {
	cmd := new(RaftCmd)
	defer func() {
		if err == nil {
			partition.saveApplyID(index)
		}
	}()
	if err = json.Unmarshal(command, cmd); err != nil {
		return
	}

	switch cmd.Op {
	case opFSMCreateInode:
		ino := NewInode(0, 0)
		if err = ino.Unmarshal(cmd.V); err != nil {
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

	klog.Infof(fmt.Sprintf("[apply raft sm]%+v", *cmd))
	return
}

func (partition *PartitionFSM) ApplyMemberChange(confChange *raftproto.ConfChange, index uint64) (interface{}, error) {
	panic("implement me")
}

func (partition *PartitionFSM) Snapshot() (raftproto.Snapshot, error) {
	panic("implement me")
}

func (partition *PartitionFSM) ApplySnapshot(peers []raftproto.Peer, iter raftproto.SnapIterator) error {
	panic("implement me")
}

func (partition *PartitionFSM) HandleFatalEvent(err *raft.FatalError) {

}

func (partition *PartitionFSM) HandleLeaderChange(leader uint64) {

}

// Put INFO: 提交 key-value 到 状态机，@see MetaPartitionFSM.Apply()
func (partition *PartitionFSM) Put(key, value interface{}) (interface{}, error) {
	raftCmd := &RaftCmd{
		Op: key.(uint32),
		K:  strconv.FormatUint(key.(uint64), 10),
		V:  value.([]byte),
	}
	cmd, _ := json.Marshal(raftCmd)
	return partition.raftPartition.Submit(cmd)
}

func (partition *PartitionFSM) Get(key interface{}) (interface{}, error) {
	panic("implement me")
}

func (partition *PartitionFSM) Del(key interface{}) (interface{}, error) {
	panic("implement me")
}

func (partition *PartitionFSM) saveApplyID(applyId uint64) {
	atomic.StoreUint64(&partition.applyID, applyId)
}

func (partition *PartitionFSM) ChangeMember(changeType raftproto.ConfChangeType, peer raftproto.Peer, context []byte) (resp interface{}, err error) {
	panic("implement me")
}

func (partition *PartitionFSM) IsLeader() (leaderAddr string, isLeader bool) {
	panic("implement me")
}

func (partition *PartitionFSM) LoadSnapshotSign(p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *PartitionFSM) DeleteRaft() error {
	panic("implement me")
}

func (partition *PartitionFSM) IsExistPeer(peer raftproto.Peer) bool {
	panic("implement me")
}

func (partition *PartitionFSM) TryToLeader(groupID uint64) error {
	panic("implement me")
}

func (partition *PartitionFSM) CanRemoveRaftMember(peer raftproto.Peer) error {
	panic("implement me")
}
