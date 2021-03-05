package master

import (
	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
	"k8s-lx1036/k8s/storage/dfs/pkg/raftstore"
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
