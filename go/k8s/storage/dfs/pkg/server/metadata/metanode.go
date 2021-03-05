package metadata

import (
	"k8s-lx1036/k8s/storage/dfs/pkg/config"
	"k8s-lx1036/k8s/storage/dfs/pkg/raftstore"
	"sync"
	"sync/atomic"
)

const (
	StateStandby uint32 = iota
	StateStart
	StateRunning
	StateShutdown
	StateStopped
)

// The MetaNode manages the dentry and inode information of the meta partitions on a meta node.
// The data consistency is ensured by Raft.
type MetaNode struct {
	nodeId            uint64
	listen            string
	metadataDir       string // root dir of the metaNode
	raftDir           string // root dir of the raftStore log
	metadataManager   MetadataManager
	localAddr         string
	clusterId         string
	raftStore         raftstore.RaftStore
	raftHeartbeatPort string
	raftReplicatePort string
	httpStopC         chan uint8
	state             uint32
	wg                sync.WaitGroup
}

// Start starts up the meta node with the specified configuration.
//  1. Start and load each meta partition from the snapshot.
//  2. Restore raftStore fsm of each meta node range.
//  3. Start server and accept connection from the master and clients.
func (m *MetaNode) Start(cfg *config.Config) (err error) {
	if atomic.CompareAndSwapUint32(&m.state, StateStandby, StateStart) {
		defer func() {
			var newState uint32
			if err != nil {
				newState = StateStandby
			} else {
				newState = StateRunning
			}
			atomic.StoreUint32(&m.state, newState)
		}()
		if err = m.onStart(cfg); err != nil {
			return
		}
		m.wg.Add(1)
	}

	return
}

func (m *MetaNode) Shutdown() {
	panic("implement me")
}

func (m *MetaNode) Sync() {
	panic("implement me")
}

// NewServer creates a new meta node instance.
func NewServer() *MetaNode {
	return &MetaNode{}
}
