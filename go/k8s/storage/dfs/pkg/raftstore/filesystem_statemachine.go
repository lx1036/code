package raftstore

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/tecbot/gorocksdb"
	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
	"k8s.io/klog/v2"
)

const (
	applied = "applied"
)

const (
	opSyncAddMetaNode          uint32 = 0x01
	opSyncAddVol               uint32 = 0x02
	opSyncAddMetaPartition     uint32 = 0x03
	opSyncUpdateMetaPartition  uint32 = 0x04
	opSyncDeleteMetaNode       uint32 = 0x05
	opSyncAllocMetaPartitionID uint32 = 0x06
	opSyncAllocCommonID        uint32 = 0x07
	opSyncPutCluster           uint32 = 0x08
	opSyncUpdateVol            uint32 = 0x09
	opSyncDeleteVol            uint32 = 0x0A
	opSyncDeleteMetaPartition  uint32 = 0x0B
	opSyncAddNodeSet           uint32 = 0x0C
	opSyncUpdateNodeSet        uint32 = 0x0D
	opSyncBatchPut             uint32 = 0x0E
	opSyncAddBucket            uint32 = 0x0F
	opSyncUpdateBucket         uint32 = 0x10
	opSyncDeleteBucket         uint32 = 0x11
	opSyncAddVolMountClient    uint32 = 0x12
	opSyncUpdateVolMountClient uint32 = 0x13
	opSyncDeleteVolMountClient uint32 = 0x14
)

const (
	GroupID = 1
)

// RaftCmd defines the Raft commands.
type RaftCmd struct {
	Op uint32 `json:"op"`
	K  string `json:"k"`
	V  []byte `json:"v"`
}

type raftLeaderChangeHandler func(leader uint64)

type raftPeerChangeHandler func(confChange *proto.ConfChange) (err error)

type raftApplySnapshotHandler func()

// FilesystemStateMachine represents the finite state machine of a metadata partition
type FilesystemStateMachine struct {
	store               *RocksDBStore
	rs                  *raft.RaftServer
	applied             uint64
	retainLogs          uint64
	leaderChangeHandler raftLeaderChangeHandler
	peerChangeHandler   raftPeerChangeHandler
	snapshotHandler     raftApplySnapshotHandler
}

// Corresponding to the LeaderChange interface in Raft library.
func (filesystemStateMachine *FilesystemStateMachine) registerLeaderChangeHandler(handler raftLeaderChangeHandler) {
	filesystemStateMachine.leaderChangeHandler = handler
}

// Corresponding to the PeerChange interface in Raft library.
func (filesystemStateMachine *FilesystemStateMachine) registerPeerChangeHandler(handler raftPeerChangeHandler) {
	filesystemStateMachine.peerChangeHandler = handler
}

// Corresponding to the ApplySnapshot interface in Raft library.
func (filesystemStateMachine *FilesystemStateMachine) registerApplySnapshotHandler(handler raftApplySnapshotHandler) {
	filesystemStateMachine.snapshotHandler = handler
}

// Get implements the interface of raft.StateMachine
func (filesystemStateMachine *FilesystemStateMachine) Get(key interface{}) (interface{}, error) {
	return filesystemStateMachine.store.Get(key)
}

// Put implements the interface of raft.StateMachine
func (filesystemStateMachine *FilesystemStateMachine) Put(key, val interface{}) (interface{}, error) {
	return filesystemStateMachine.store.Put(key, val, true)
}

// Del implements the interface of raft.StateMachine
func (filesystemStateMachine *FilesystemStateMachine) Del(key interface{}) (interface{}, error) {
	return filesystemStateMachine.store.Del(key, true)
}

// Apply implements the interface of raft.StateMachine
func (filesystemStateMachine *FilesystemStateMachine) Apply(command []byte, index uint64) (resp interface{}, err error) {
	cmd := new(RaftCmd)
	if err = json.Unmarshal(command, cmd); err != nil {
		klog.Errorf("action[fsmApply],unmarshal data:%v, err:%v", command, err.Error())
		panic(err)
	}
	klog.Infof("action[fsmApply],cmd.op[%v],cmd.K[%v],cmd.V[%v]", cmd.Op, cmd.K, string(cmd.V))
	cmdMap := make(map[string][]byte)
	if cmd.Op != opSyncBatchPut {
		cmdMap[cmd.K] = cmd.V
		cmdMap[applied] = []byte(strconv.FormatUint(uint64(index), 10))
	} else {
		nestedCmdMap := make(map[string]*RaftCmd)
		if err = json.Unmarshal(cmd.V, &nestedCmdMap); err != nil {
			klog.Errorf("action[fsmApply],unmarshal nested cmd data:%v, err:%v", command, err.Error())
			panic(err)
		}
		for cmdK, cmd := range nestedCmdMap {
			klog.Infof("action[fsmApply],cmd.op[%v],cmd.K[%v],cmd.V[%v]", cmd.Op, cmd.K, string(cmd.V))
			cmdMap[cmdK] = cmd.V
		}
		cmdMap[applied] = []byte(strconv.FormatUint(uint64(index), 10))
	}
	switch cmd.Op {
	case opSyncDeleteMetaNode, opSyncDeleteVol, opSyncDeleteMetaPartition, opSyncDeleteBucket, opSyncDeleteVolMountClient:
		if err = filesystemStateMachine.delKeyAndPutIndex(cmd.K, cmdMap); err != nil {
			panic(err)
		}
	default:
		if err = filesystemStateMachine.batchPut(cmdMap); err != nil {
			panic(err)
		}
	}
	filesystemStateMachine.applied = index
	if filesystemStateMachine.applied > 0 && (filesystemStateMachine.applied%filesystemStateMachine.retainLogs) == 0 {
		klog.Warningf("action[Apply],truncate raft log,retainLogs[%v],index[%v]", filesystemStateMachine.retainLogs, filesystemStateMachine.applied)
		filesystemStateMachine.rs.Truncate(GroupID, filesystemStateMachine.applied)
	}
	return
}

// ApplyMemberChange implements the interface of raft.StateMachine
func (filesystemStateMachine *FilesystemStateMachine) ApplyMemberChange(confChange *proto.ConfChange, index uint64) (interface{}, error) {
	var err error
	if filesystemStateMachine.peerChangeHandler != nil {
		err = filesystemStateMachine.peerChangeHandler(confChange)
	}
	return nil, err
}

// Snapshot implements the interface of raft.StateMachine
func (filesystemStateMachine *FilesystemStateMachine) Snapshot() (proto.Snapshot, error) {
	snapshot := filesystemStateMachine.store.RocksDBSnapshot()
	iterator := filesystemStateMachine.store.Iterator(snapshot)
	iterator.SeekToFirst()
	return &MetadataSnapshot{
		applied:  filesystemStateMachine.applied,
		snapshot: snapshot,
		fsm:      filesystemStateMachine,
		iterator: iterator,
	}, nil
}

// ApplySnapshot implements the interface of raft.StateMachine
func (filesystemStateMachine *FilesystemStateMachine) ApplySnapshot(peers []proto.Peer, iterator proto.SnapIterator) error {
	klog.Infof(fmt.Sprintf("action[ApplySnapshot] begin,applied[%v]", filesystemStateMachine.applied))
	var data []byte
	var err error
	for {
		if data, err = iterator.Next(); err != nil {
			break
		}
		cmd := &RaftCmd{}
		if err = json.Unmarshal(data, cmd); err != nil {
			klog.Errorf("action[ApplySnapshot] failed,err:%v", err)
			return err
		}
		if _, err = filesystemStateMachine.store.Put(cmd.K, cmd.V, true); err != nil {
			klog.Errorf("action[ApplySnapshot] failed,err:%v", err)
			return err
		}
	}
	if err != io.EOF {
		klog.Errorf("action[ApplySnapshot] failed,err:%v", err)
		return err
	}

	if filesystemStateMachine.snapshotHandler != nil {
		filesystemStateMachine.snapshotHandler()
	}

	klog.Infof(fmt.Sprintf("action[ApplySnapshot] success,applied[%v]", filesystemStateMachine.applied))
	return nil
}

// HandleFatalEvent implements the interface of raft.StateMachine
func (filesystemStateMachine *FilesystemStateMachine) HandleFatalEvent(err *raft.FatalError) {
	panic(err.Err)
}

// HandleLeaderChange implements the interface of raft.StateMachine
func (filesystemStateMachine *FilesystemStateMachine) HandleLeaderChange(leader uint64) {
	if filesystemStateMachine.leaderChangeHandler != nil {
		go filesystemStateMachine.leaderChangeHandler(leader)
	}
}

func (filesystemStateMachine *FilesystemStateMachine) delKeyAndPutIndex(key string, cmdMap map[string][]byte) (err error) {
	return filesystemStateMachine.store.DeleteKeyAndPutIndex(key, cmdMap, true)
}

func (filesystemStateMachine *FilesystemStateMachine) batchPut(cmdMap map[string][]byte) (err error) {
	return filesystemStateMachine.store.BatchPut(cmdMap, true)
}

func (filesystemStateMachine *FilesystemStateMachine) Restore() {
	value, err := filesystemStateMachine.Get(applied)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore applied err:%v", err))
	}
	byteValues := value.([]byte)
	if len(byteValues) == 0 {
		filesystemStateMachine.applied = 0
		return
	}
	applied, err := strconv.ParseUint(string(byteValues), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore applied,err:%v ", err))
	}
	filesystemStateMachine.applied = applied
}

func (filesystemStateMachine *FilesystemStateMachine) GetApply() uint64 {
	return filesystemStateMachine.applied
}

func NewFilesystemStateMachine(store *RocksDBStore, retainsLog uint64, rs *raft.RaftServer) *FilesystemStateMachine {
	return &FilesystemStateMachine{
		store:      store,
		rs:         rs,
		retainLogs: retainsLog,
	}
}

// MetadataSnapshot represents the snapshot of a meta partition
type MetadataSnapshot struct {
	fsm      *FilesystemStateMachine
	applied  uint64
	snapshot *gorocksdb.Snapshot
	iterator *gorocksdb.Iterator
}

func (ms *MetadataSnapshot) Next() ([]byte, error) {
	panic("implement me")
}

// ApplyIndex implements the Snapshot interface
func (ms *MetadataSnapshot) ApplyIndex() uint64 {
	return ms.applied
}

func (ms *MetadataSnapshot) Close() {
	ms.fsm.store.ReleaseSnapshot(ms.snapshot)
}
