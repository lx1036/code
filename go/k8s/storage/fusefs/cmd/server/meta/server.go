package meta

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
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

// Configuration keys
const (
	cfgLocalIP           = "localIP"
	cfgListen            = "listen"
	cfgMetadataDir       = "metadataDir"
	cfgRaftDir           = "raftDir"
	cfgMasterAddrs       = "masterAddrs"
	cfgRaftHeartbeatPort = "raftHeartbeatPort"
	cfgRaftReplicaPort   = "raftReplicaPort"
	cfgTotalMem          = "totalMem"

	profPort = "profPort"
)

var (
	masterHelper   util.MasterHelper
	configTotalMem uint64
)

type RaftCmd struct {
	Op uint32 `json:"op"`
	K  string `json:"k"`
	V  []byte `json:"v"`
}

type Config struct {
	IP          string   `json:"ip"`
	Port        int      `json:"port"`
	TotalMem    uint64   `json:"totalMem"`
	MasterAddrs []string `json:"masterAddrs"`

	// raft
	StoreDir          string `json:"storeDir"`
	WalDir            string `json:"walDir"`
	RaftHeartbeatPort int    `json:"raftHeartbeatPort"`
	RaftReplicaPort   int    `json:"raftReplicaPort"`
	RetainLogs        uint64 `json:"retainLogs"`
}

// The Server manages the dentry and inode information of the meta partitions on a meta node.
// The data consistency is ensured by Raft.
type Server struct {
	ip           string
	nodeId       uint64
	clusterName  string
	port         int
	masterAddrs  []string
	masterLeader string
	state        uint32

	cluster *Cluster

	storeDir          string
	walDir            string
	raftHeartbeatPort int
	raftReplicatePort int
	retainLogs        uint64
	raftStore         *raftstore.RaftStore
}

func NewServer(config Config) *Server {
	return &Server{
		ip:           config.IP,
		port:         config.Port,
		masterAddrs:  config.MasterAddrs,
		masterLeader: config.MasterAddrs[0], // TODO: 暂时选择第一个作为 leader address

		storeDir:          config.StoreDir,
		walDir:            config.WalDir,
		raftHeartbeatPort: config.RaftHeartbeatPort,
		raftReplicatePort: config.RaftReplicaPort,
		retainLogs:        config.RetainLogs,
	}
}

// Start starts up the meta node with the specified configuration.
//  1. Start and load each meta partition from the snapshot.
//  2. Restore raftStore fsm of each meta node range.
//  3. Start server and accept connection from the master and clients.
func (server *Server) Start() (err error) {
	if server.raftStore, err = raftstore.NewRaftStore(&raftstore.Config{
		NodeID:            server.nodeId,
		IPAddr:            server.ip,
		HeartbeatPort:     server.raftHeartbeatPort,
		ReplicaPort:       server.raftReplicatePort,
		RaftPath:          server.walDir,
		NumOfLogsToRetain: server.retainLogs,
		TickInterval:      1000, // 1s
		ElectionTick:      5,    // [5 * 1s, 2 * 5 * 1s)
	}); err != nil {
		return err
	}

	// INFO: 向 master 注册 meta, POST /metaNode/add
	server.nodeId = 1 // for debug in local
	if err = server.registerToMaster(); err != nil {
		return err
	}

	// INFO: 用来处理来自 client 的请求，并把数据写到 raft log 里
	server.cluster = NewCluster(MetadataManagerConfig{
		NodeID:    server.nodeId,
		StoreDir:  server.storeDir,
		RaftStore: server.raftStore,
	})
	if err = server.cluster.Start(); err != nil {
		return err
	}

	// INFO: goroutine 启动 TCP server，被 client 调用!!!
	return server.startTCPServer()
}

// INFO: 向 master 注册 meta, POST /metaNode, @see server/master/http_server.go
func (server *Server) registerToMaster() error {
	resp, err := http.Get(fmt.Sprintf("%s:%d%s", server.masterLeader, server.port, "/cluster/info"))
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	clusterInfo := &proto.ClusterInfo{}
	_ = json.Unmarshal(data, clusterInfo)
	server.clusterName = clusterInfo.Cluster

	// register meta node
	reqParam := make(map[string]string)
	reqParam["addr"] = fmt.Sprintf("%s:%d", server.ip, server.port)
	url := fmt.Sprintf("%s:%d?addr=%s", server.masterLeader, server.port, fmt.Sprintf("%s:%d", server.ip, server.port))
	resp2, err := http.Post(url, "application/json", nil)
	metaNodeID, err := ioutil.ReadAll(resp2.Body)
	defer resp2.Body.Close()
	nodeIDStr := strings.TrimSpace(string(metaNodeID))
	if len(nodeIDStr) == 0 {
		return fmt.Errorf("[register] master respond empty body")
	}
	server.nodeId, err = strconv.ParseUint(nodeIDStr, 10, 64)
	return err
}

func (server *Server) startTCPServer() (err error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.ip, server.port))
	if err != nil {
		klog.Fatalf(fmt.Sprintf("tcp err:%v", err))
		return
	}
	go func() {
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				klog.Fatalf(fmt.Sprintf("tcp err:%v", err))
			}
			go server.serveConn(conn)
		}
	}()
	klog.Infof("start tcp server...")
	return
}

// INFO: 从 tcp connection 读数据
func (server *Server) serveConn(conn net.Conn) {
	defer conn.Close()
	c := conn.(*net.TCPConn)
	c.SetKeepAlive(true)
	remoteAddr := conn.RemoteAddr().String()
	for {
		p := &proto.Packet{}
		if err := p.ReadFromConn(conn, proto.NoReadDeadlineTime); err != nil {
			if err != io.EOF {
				klog.Errorf("serve Server remote[%v] %v error: %v", remoteAddr, p.GetUniqueLogId(), err)
			}
			return
		}

		// Start a goroutine for packet handling. Do not block connection read goroutine.
		go func() {
			if err := server.cluster.HandleMetadataOperation(conn, p, remoteAddr); err != nil {
				klog.Errorf("metadata handle connection err: %v", err)
				return
			}
		}()
	}
}

func (server *Server) Stop() {
	server.cluster.Stop()
	server.raftStore.Stop()
}
