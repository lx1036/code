package meta

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/fusefs/pkg/config"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s-lx1036/k8s/storage/fusefs/pkg/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"

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
	metadataManager   *metadataManager
	localAddr         string
	clusterId         string
	raftStore         raftstore.RaftStore
	raftHeartbeatPort string
	raftReplicatePort string
	httpStopC         chan uint8
	state             uint32
	wg                sync.WaitGroup
}

func NewServer() *MetaNode {
	return &MetaNode{}
}

// Start starts up the meta node with the specified configuration.
//  1. Start and load each meta partition from the snapshot.
//  2. Restore raftStore fsm of each meta node range.
//  3. Start server and accept connection from the master and clients.
func (m *MetaNode) Start(cfg *config.Config) error {
	var err error
	if err = m.parseConfig(cfg); err != nil {
		return err
	}

	// INFO: 向 master 注册 meta, POST /metaNode/add
	m.nodeId = 1 // for debug in local
	/*if err = m.register(); err != nil {
		return err
	}*/

	// INFO: 启动 raft，等待 meta manager 来提交 raft log
	if err = m.startRaftServer(); err != nil {
		return err
	}

	// TODO: http handler 后续实现
	/*if err = m.registerAPIHandler(); err != nil {
		return err
	}*/
	// INFO: 用来处理来自 client 的请求，并把数据写到 raft log 里
	if err = m.startMetaManager(); err != nil {
		return err
	}

	// check local partition compare with master ,if lack,then not start
	/*if err = m.checkLocalPartitionMatchWithMaster(); err != nil {
		klog.Error(err)
		return err
	}*/

	// INFO: goroutine 启动 TCP server，监听在 9021 port，被 client 调用!!!
	if err = m.startTCPServer(); err != nil {
		return err
	}

	m.wg.Add(1)

	return nil
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

// INFO: GET masterAddrs[len(masterAddrs)-1]:9500/admin/getIp
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

// INFO: tcp 监听在 9021 port，没有走 http 协议
func (m *MetaNode) startTCPServer() (err error) {
	m.httpStopC = make(chan uint8)
	ln, err := net.Listen("tcp", ":"+m.listen)
	if err != nil {
		return
	}
	go func(stopC chan uint8) {
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			select {
			case <-stopC:
				return
			default:
			}
			if err != nil {
				continue
			}
			go m.serveConn(conn, stopC)
		}
	}(m.httpStopC)
	klog.Infof("start server over...")
	return
}

func (m *MetaNode) stopTCPServer() {
	if m.httpStopC != nil {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf(fmt.Sprintf("action[StopTcpServer],err:%v", r))
			}
		}()
		close(m.httpStopC)
	}
}

// INFO: 从 tcp connection 读数据
func (m *MetaNode) serveConn(conn net.Conn, stopC chan uint8) {
	defer conn.Close()
	c := conn.(*net.TCPConn)
	c.SetKeepAlive(true)
	c.SetNoDelay(true)
	remoteAddr := conn.RemoteAddr().String()
	for {
		select {
		case <-stopC:
			return
		default:
		}
		p := &proto.Packet{}
		if err := p.ReadFromConn(conn, proto.NoReadDeadlineTime); err != nil {
			if err != io.EOF {
				klog.Errorf("serve MetaNode remote[%v] %v error: %v", remoteAddr, p.GetUniqueLogId(), err)
			}
			return
		}

		// Start a goroutine for packet handling. Do not block connection read goroutine.
		go func() {
			if err := m.metadataManager.HandleMetadataOperation(conn, p, remoteAddr); err != nil {
				klog.Errorf("serve operatorPkg: %v", err)
				return
			}
		}()
	}
}

// INFO: check meta local partition 和 master partition 数据是否一致
//  主要是看存储在 master 上的 partition id，是否在本地有对应的 partition
func (m *MetaNode) checkLocalPartitionMatchWithMaster() (err error) {
	params := make(map[string]string)
	params["addr"] = m.localAddr + ":" + m.listen
	data, err := masterHelper.Request(http.MethodGet, proto.GetMetaNode, params, nil)
	if err != nil {
		return fmt.Errorf("checkLocalPartitionMatchWithMaster error %v", err)
	}

	type MetaNodeInfo struct {
		Addr                      string
		PersistenceMetaPartitions []uint64
	}
	metaNodeInfo := new(MetaNodeInfo)
	if err = json.Unmarshal(data, metaNodeInfo); err != nil {
		return fmt.Errorf("checkLocalPartitionMatchWithMaster jsonUnmarsh failed %v", err)
	}

	if len(metaNodeInfo.PersistenceMetaPartitions) == 0 {
		return
	}
	lackPartitions := make([]uint64, 0)
	for _, partitionID := range metaNodeInfo.PersistenceMetaPartitions {
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

func (m *MetaNode) startMetaManager() (err error) {
	if _, err = os.Stat(m.metadataDir); err != nil {
		if err = os.MkdirAll(m.metadataDir, 0755); err != nil {
			return
		}
	}
	m.metadataManager = NewMetadataManager(MetadataManagerConfig{
		NodeID:    m.nodeId,
		RootDir:   m.metadataDir,
		RaftStore: m.raftStore,
	})
	if err = m.metadataManager.Start(); err == nil {
		klog.Infof("[startMetaManager] manager start finish.")
	}
	return
}

func (m *MetaNode) stopMetaManager() {
	if m.metadataManager != nil {
		m.metadataManager.Stop()
	}
}

func (m *MetaNode) Shutdown() {
	// shutdown node and release the resource
	m.stopTCPServer()
	m.stopMetaManager()
	m.stopRaftServer()
	m.wg.Done()
}

func (m *MetaNode) Wait() {
	if atomic.LoadUint32(&m.state) == StateRunning {
		m.wg.Wait()
	}
}
