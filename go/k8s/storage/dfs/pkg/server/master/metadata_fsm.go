package master

import (
	"fmt"
	"strconv"

	"k8s-lx1036/k8s/storage/dfs/pkg/raftstore"

	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
)

const (
	applied = "applied"
)

type raftLeaderChangeHandler func(leader uint64)

type raftPeerChangeHandler func(confChange *proto.ConfChange) (err error)

type raftApplySnapshotHandler func()

// MetadataFsm represents the finite state machine of a metadata partition
type MetadataFsm struct {
	store               *raftstore.RocksDBStore
	rs                  *raft.RaftServer
	applied             uint64
	retainLogs          uint64
	leaderChangeHandler raftLeaderChangeHandler
	peerChangeHandler   raftPeerChangeHandler
	snapshotHandler     raftApplySnapshotHandler
}

// Corresponding to the LeaderChange interface in Raft library.
func (mf *MetadataFsm) registerLeaderChangeHandler(handler raftLeaderChangeHandler) {
	mf.leaderChangeHandler = handler
}

// Corresponding to the PeerChange interface in Raft library.
func (mf *MetadataFsm) registerPeerChangeHandler(handler raftPeerChangeHandler) {
	mf.peerChangeHandler = handler
}

// Corresponding to the ApplySnapshot interface in Raft library.
func (mf *MetadataFsm) registerApplySnapshotHandler(handler raftApplySnapshotHandler) {
	mf.snapshotHandler = handler
}

// Get implements the interface of raft.StateMachine
func (mf *MetadataFsm) Get(key interface{}) (interface{}, error) {
	return mf.store.Get(key)
}

func (mf *MetadataFsm) restore() {
	value, err := mf.Get(applied)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore applied err:%v", err))
	}
	byteValues := value.([]byte)
	if len(byteValues) == 0 {
		mf.applied = 0
		return
	}
	applied, err := strconv.ParseUint(string(byteValues), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore applied,err:%v ", err))
	}
	mf.applied = applied
}

func newMetadataFsm(store *raftstore.RocksDBStore, retainsLog uint64, rs *raft.RaftServer) (fsm *MetadataFsm) {
	return &MetadataFsm{
		store:      store,
		rs:         rs,
		retainLogs: retainsLog,
	}
}
