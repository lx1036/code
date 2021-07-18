package meta

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	masterHelper   util.MasterHelper
	configTotalMem uint64
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

type MetaNodeInfo struct {
	Addr                      string
	PersistenceMetaPartitions []uint64
}

func (m *MetaNode) checkLocalPartitionMatchWithMaster() (err error) {
	params := make(map[string]string)
	params["addr"] = m.localAddr + ":" + m.listen
	data, err := masterHelper.Request(http.MethodGet, proto.GetMetaNode, params, nil)
	if err != nil {
		return fmt.Errorf("checkLocalPartitionMatchWithMaster error %v", err)
	}

	minfo := new(MetaNodeInfo)
	if err = json.Unmarshal(data, minfo); err != nil {
		return fmt.Errorf("checkLocalPartitionMatchWithMaster jsonUnmarsh failed %v", err)
	}

	if len(minfo.PersistenceMetaPartitions) == 0 {
		return
	}
	lackPartitions := make([]uint64, 0)
	for _, partitionID := range minfo.PersistenceMetaPartitions {
		_, err := m.metadataManager.GetPartition(partitionID)
		if err != nil {
			lackPartitions = append(lackPartitions, partitionID)
		}
	}
	if len(lackPartitions) == 0 {
		return
	}

	return fmt.Errorf("LackPartitions %v on metanode %v,metanode cannot start", lackPartitions, m.localAddr+":"+m.listen)
}

// INFO: 向 master 注册 meta, POST /metaNode/add
func (m *MetaNode) register() (err error) {
	reqParam := make(map[string]string)
	clusterInfo, err := getClusterInfo()
	if err != nil {
		klog.Errorf("[register] %s", err.Error())
		return err
	}

	if m.localAddr == "" {
		m.localAddr = clusterInfo.Ip
	}
	m.clusterId = clusterInfo.Cluster
	reqParam["addr"] = m.localAddr + ":" + m.listen

	respBody, err := masterHelper.Request("POST", proto.AddMetaNode, reqParam, nil)
	if err != nil {

	}
	nodeIDStr := strings.TrimSpace(string(respBody))
	if nodeIDStr == "" {
		return fmt.Errorf("[register] master respond empty body")
	}
	m.nodeId, err = strconv.ParseUint(nodeIDStr, 10, 64)
	if err != nil {
		return err
	}

	return nil
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
	//m.stopServer()
	//m.stopMetaManager()
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

func (m *MetaNode) parseConfig(cfg *config.Config) (err error) {
	if cfg == nil {
		err = errors.New("invalid configuration")
		return
	}
	m.localAddr = cfg.GetString(cfgLocalIP)
	m.listen = cfg.GetString(cfgListen)
	m.metadataDir = cfg.GetString(cfgMetadataDir)
	m.raftDir = cfg.GetString(cfgRaftDir)
	m.raftHeartbeatPort = cfg.GetString(cfgRaftHeartbeatPort)
	m.raftReplicatePort = cfg.GetString(cfgRaftReplicaPort)
	configTotalMem, _ = strconv.ParseUint(cfg.GetString(cfgTotalMem), 10, 64)
	if configTotalMem == 0 {
		return fmt.Errorf("bad totalMem config,Recommended to be configured as 80 percent of physical machine memory")
	}
	if m.metadataDir == "" {
		return fmt.Errorf("bad metadataDir config")
	}
	if m.listen == "" {
		return fmt.Errorf("bad listen config")
	}
	if m.raftDir == "" {
		return fmt.Errorf("bad raftDir config")
	}
	if m.raftHeartbeatPort == "" {
		return fmt.Errorf("bad raftHeartbeatPort config")
	}
	if m.raftReplicatePort == "" {
		return fmt.Errorf("bad cfgRaftReplicaPort config")
	}

	// INFO: 向 master 中注册 meta，没有 tcp 请求
	addrs := cfg.GetArray(cfgMasterAddrs)
	masterHelper = util.NewMasterHelper()
	for _, addr := range addrs {
		masterHelper.AddNode(addr.(string))
	}
	//err = m.validConfig()
	return
}

// NewServer creates a new meta node instance.
func NewServer() *MetaNode {
	return &MetaNode{}
}
