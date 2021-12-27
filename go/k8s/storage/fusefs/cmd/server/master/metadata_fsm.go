package master

import (
	"encoding/json"
	"fmt"
	"io"
	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"
	"strconv"

	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
	"k8s.io/klog/v2"
)

const (
	applied = "applied"
)

type raftLeaderChangeHandler func(leader uint64)

type raftPeerChangeHandler func(confChange *proto.ConfChange) (err error)

type raftApplySnapshotHandler func()

// MetadataFsm INFO: MetadataFsm 是一个 state machine
type MetadataFsm struct {
	store               *raftstore.BoltDBStore
	raftServer          *raft.RaftServer
	applied             uint64
	retainLogs          uint64
	leaderChangeHandler raftLeaderChangeHandler
	peerChangeHandler   raftPeerChangeHandler
	snapshotHandler     raftApplySnapshotHandler
}

// Corresponding to the LeaderChange interface in Raft library.
func (metadataFsm *MetadataFsm) registerLeaderChangeHandler(handler raftLeaderChangeHandler) {
	metadataFsm.leaderChangeHandler = handler
}

// Corresponding to the PeerChange interface in Raft library.
func (metadataFsm *MetadataFsm) registerPeerChangeHandler(handler raftPeerChangeHandler) {
	metadataFsm.peerChangeHandler = handler
}

// Corresponding to the ApplySnapshot interface in Raft library.
func (metadataFsm *MetadataFsm) registerApplySnapshotHandler(handler raftApplySnapshotHandler) {
	metadataFsm.snapshotHandler = handler
}

// Get implements the interface of raft.StateMachine
func (metadataFsm *MetadataFsm) Get(key []byte) (interface{}, error) {
	return metadataFsm.store.Get(key)
}

// Put implements the interface of raft.StateMachine
func (metadataFsm *MetadataFsm) Put(key, val []byte) error {
	return metadataFsm.store.Put(key, val)
}

// Del implements the interface of raft.StateMachine
func (metadataFsm *MetadataFsm) Del(key []byte) error {
	return metadataFsm.store.Del(key)
}

// Apply implements the interface of raft.StateMachine
func (metadataFsm *MetadataFsm) Apply(command []byte, index uint64) (resp interface{}, err error) {
	cmd := new(RaftCmd)
	if err = json.Unmarshal(command, cmd); err != nil {
		klog.Errorf(fmt.Sprintf("apply fsm for cmd:%s, index:%d, err:%v", command, index, err))
		return
	}
	klog.Infof(fmt.Sprintf("apply fsm for cmd:%+v, index:%d", *cmd, index))
	/*cmdMap := make(map[string][]byte)
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
		if err = metadataFsm.delKeyAndPutIndex(cmd.K, cmdMap); err != nil {
			panic(err)
		}
	default:
		if err = metadataFsm.batchPut(cmdMap); err != nil {
			panic(err)
		}
	}
	metadataFsm.applied = index
	if metadataFsm.applied > 0 {
		//if metadataFsm.applied > 0 && (metadataFsm.applied%metadataFsm.retainLogs) == 0 {
		klog.Warningf("action[Apply],truncate raft log,retainLogs[%v],index[%v]", metadataFsm.retainLogs, metadataFsm.applied)
		metadataFsm.raftServer.Truncate(GroupID, metadataFsm.applied)
	}*/
	return
}

// ApplyMemberChange implements the interface of raft.StateMachine
func (metadataFsm *MetadataFsm) ApplyMemberChange(confChange *proto.ConfChange, index uint64) (interface{}, error) {
	var err error
	if metadataFsm.peerChangeHandler != nil {
		err = metadataFsm.peerChangeHandler(confChange)
	}
	return nil, err
}

// Snapshot implements the interface of raft.StateMachine
func (metadataFsm *MetadataFsm) Snapshot() (proto.Snapshot, error) {
	//snapshot := metadataFsm.store.RocksDBSnapshot()
	//iterator := metadataFsm.store.Iterator(snapshot)
	//iterator.SeekToFirst()
	return &MetadataSnapshot{
		applied: metadataFsm.applied,
		//snapshot: snapshot,
		fsm: metadataFsm,
		//iterator: iterator,
	}, nil
}

// ApplySnapshot implements the interface of raft.StateMachine
func (metadataFsm *MetadataFsm) ApplySnapshot(peers []proto.Peer, iterator proto.SnapIterator) error {
	klog.Infof(fmt.Sprintf("action[ApplySnapshot] begin,applied[%v]", metadataFsm.applied))
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
		if err = metadataFsm.store.Put([]byte(cmd.K), cmd.V); err != nil {
			klog.Errorf("action[ApplySnapshot] failed,err:%v", err)
			return err
		}
	}
	if err != io.EOF {
		klog.Errorf("action[ApplySnapshot] failed,err:%v", err)
		return err
	}

	if metadataFsm.snapshotHandler != nil {
		metadataFsm.snapshotHandler()
	}

	klog.Infof(fmt.Sprintf("action[ApplySnapshot] success,applied[%v]", metadataFsm.applied))
	return nil
}

// HandleFatalEvent implements the interface of raft.StateMachine
func (metadataFsm *MetadataFsm) HandleFatalEvent(err *raft.FatalError) {
	panic(err.Err)
}

// HandleLeaderChange implements the interface of raft.StateMachine
func (metadataFsm *MetadataFsm) HandleLeaderChange(leader uint64) {
	if metadataFsm.leaderChangeHandler != nil {
		go metadataFsm.leaderChangeHandler(leader)
	}
}

func (metadataFsm *MetadataFsm) delKeyAndPutIndex(key string, cmdMap map[string][]byte) (err error) {
	//return metadataFsm.store.DeleteKeyAndPutIndex(key, cmdMap, true)
	panic("not implemented")
}

func (metadataFsm *MetadataFsm) batchPut(cmdMap map[string][]byte) (err error) {
	//return metadataFsm.store.BatchPut(cmdMap, true)
	panic("not implemented")
}

func (metadataFsm *MetadataFsm) Restore() {
	value, err := metadataFsm.Get([]byte(applied))
	if err != nil {
		panic(fmt.Sprintf("Failed to restore applied err:%v", err))
	}
	byteValues := value.([]byte)
	if len(byteValues) == 0 {
		metadataFsm.applied = 0
		return
	}
	applied, err := strconv.ParseUint(string(byteValues), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore applied,err:%v ", err))
	}
	metadataFsm.applied = applied
}

func (metadataFsm *MetadataFsm) GetApply() uint64 {
	return metadataFsm.applied
}

func (metadataFsm *MetadataFsm) restore() {
	metadataFsm.restoreApplied()
}

func (metadataFsm *MetadataFsm) restoreApplied() {
	value, err := metadataFsm.Get([]byte(applied))
	if err != nil {
		panic(fmt.Sprintf("Failed to restore applied err:%v", err.Error()))
	}
	byteValues := value.([]byte)
	if len(byteValues) == 0 {
		metadataFsm.applied = 0
		return
	}
	restoredValues, err := strconv.ParseUint(string(byteValues), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore applied,err:%v ", err.Error()))
	}
	metadataFsm.applied = restoredValues
}

// INFO: https://github.com/tiglabs/raft/blob/master/test/memory_statemachine.go
func newMetadataFsm(store *raftstore.BoltDBStore, retainsLog uint64, raftServer *raft.RaftServer) *MetadataFsm {
	return &MetadataFsm{
		store:      store,
		raftServer: raftServer,
		retainLogs: retainsLog,
	}
}

// MetadataSnapshot represents the snapshot of a meta partition
type MetadataSnapshot struct {
	fsm     *MetadataFsm
	applied uint64
	//snapshot *gorocksdb.Snapshot
	//iterator *gorocksdb.Iterator
}

func (ms *MetadataSnapshot) Next() ([]byte, error) {
	panic("implement me")
}

// ApplyIndex implements the Snapshot interface
func (ms *MetadataSnapshot) ApplyIndex() uint64 {
	return ms.applied
}

func (ms *MetadataSnapshot) Close() {
	//ms.fsm.store.ReleaseSnapshot(ms.snapshot)
}
