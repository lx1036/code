package meta

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/sunfs/pkg/config"
	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"
	"k8s-lx1036/k8s/storage/sunfs/pkg/raftstore"
	"k8s-lx1036/k8s/storage/sunfs/pkg/util"

	"k8s.io/klog/v2"
)

const (
	StateStandby uint32 = iota
	StateStart
	StateRunning
	StateShutdown
	StateStopped
)

var (
	masterHelper util.MasterHelper
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

func (m *MetaNode) onStart(cfg *config.Config) error {
	var err error
	if err = m.parseConfig(cfg); err != nil {
		return err
	}
	if err = m.register(); err != nil {
		return err
	}
	if err = m.startRaftServer(); err != nil {
		return err
	}
	if err = m.registerAPIHandler(); err != nil {
		return err
	}
	if err = m.startMetaManager(); err != nil {
		return err
	}

	// check local partition compare with master ,if lack,then not start
	if err = m.checkLocalPartitionMatchWithMaster(); err != nil {
		klog.Error(err)
		return err
	}

	if err = m.startServer(); err != nil {
		return err
	}

	return nil
}

// 向 master 注册 meta
func (m *MetaNode) register() (err error) {
	clusterInfo, err = getClusterInfo()
	if err != nil {
		klog.Errorf("[register] %s", err.Error())
		return err
	}

	if m.localAddr == "" {
		m.localAddr = clusterInfo.Ip
	}
	m.clusterId = clusterInfo.Cluster
	reqParam["addr"] = m.localAddr + ":" + m.listen

	respBody, err = masterHelper.Request("POST", proto.AddMetaNode, reqParam, nil)
	if err != nil {

	}
	nodeIDStr := strings.TrimSpace(string(respBody))
	if nodeIDStr == "" {

	}
	m.nodeId, err = strconv.ParseUint(nodeIDStr, 10, 64)
	if err != nil {

	}

}

func (m *MetaNode) startMetaManager() (err error) {
	if _, err = os.Stat(m.metadataDir); err != nil {
		if err = os.MkdirAll(m.metadataDir, 0755); err != nil {
			return
		}
	}
	// load metadataManager
	conf := MetadataManagerConfig{
		NodeID:    m.nodeId,
		RootDir:   m.metadataDir,
		RaftStore: m.raftStore,
	}
	m.metadataManager = NewMetadataManager(conf)
	if err = m.metadataManager.Start(); err == nil {
		klog.Infof("[startMetaManager] manager start finish.")
	}
	return
}

func (m *MetaNode) Shutdown() {
	// shutdown node and release the resource
	m.stopServer()
	m.stopMetaManager()
	m.stopRaftServer()
}

func (m *MetaNode) Sync() {
	if atomic.LoadUint32(&m.state) == StateRunning {
		m.wg.Wait()
	}
}

func getClusterInfo() (*proto.ClusterInfo, error) {
	respBody, err := masterHelper.Request("GET", proto.AdminGetIP, nil, nil)
	if err != nil {
		return nil, err
	}
	cInfo := &proto.ClusterInfo{}
	if err = json.Unmarshal(respBody, cInfo); err != nil {
		return nil, err
	}
	return cInfo, nil
}

// NewServer creates a new meta node instance.
func NewServer() *MetaNode {
	return &MetaNode{}
}
