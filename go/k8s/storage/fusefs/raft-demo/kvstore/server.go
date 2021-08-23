package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
	"github.com/tiglabs/raft/storage/wal"
	"k8s.io/klog/v2"
)

// DefaultClusterID the default cluster id, we have only one raft cluster
const DefaultClusterID = 1

// DefaultRequestTimeout default request timeout
const DefaultRequestTimeout = time.Second * 3

// CmdType command type
type CmdType int

const (
	// CmdQuorumGet quorum get a key
	CmdQuorumGet CmdType = 1
	// CmdPut put key value
	CmdPut CmdType = 2
	// CmdDelete delete a key
	CmdDelete CmdType = 3
)

// Command a raft op command
type Command struct {
	OP    CmdType `json:"op"`
	Key   []byte  `json:"k"`
	Value []byte  `json:"v"`
}

func (c *Command) String() string {
	switch c.OP {
	case CmdQuorumGet:
		return fmt.Sprintf("QuorumGet %v", string(c.Key))
	case CmdPut:
		return fmt.Sprintf("Put %s %s", string(c.Key), string(c.Value))
	case CmdDelete:
		return fmt.Sprintf("Delete %s", string(c.Key))
	default:
		return "<Invalid>"
	}
}

// Server the kv server
// TODO Server 实现了 StateMachine 接口
type Server struct {
	cfg    *Config
	nodeID uint64       // self node id
	node   *ClusterNode // self node

	hs         *http.Server
	raftServer *raft.RaftServer
	db         *Store

	leader uint64
}

func (server *Server) startHTTPServer() {
	router := mux.NewRouter()
	router.HandleFunc("/kvs/{key}", server.Get).Methods("GET")
	router.HandleFunc("/kvs/{key}", server.Put).Methods("PUT")
	router.HandleFunc("/kvs/{key}", server.Delete).Methods("DELETE")

	addr := fmt.Sprintf(":%d", server.node.HTTPPort)
	server.hs = &http.Server{
		Addr:    addr,
		Handler: router,
	}
	err := server.hs.ListenAndServe()
	if err != nil {
		klog.Fatalf("listen http on %v failed: %v", addr, err)
	}

	klog.Infof("http start listen on %v", addr)
}

// Get get a key
func (server *Server) Get(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	if err := r.ParseForm(); err != nil {
		klog.Errorf("call ParseForm failed: %v", err)
		return
	}

	level := r.Form.Get("level")
	klog.Infof("[Get]level %s", level)
	switch level {
	case "log":
		server.process(w, CmdQuorumGet, []byte(key), nil)
		return
	case "index":
		server.getByReadIndex(w, key)
		return
	default:
		value, err := server.db.Get([]byte(key))
		if err != nil {
			klog.Errorf(fmt.Sprintf("[Get]get key %s with err %v", key, err))
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.Write(value)
		}
	}
}

// Put put
func (server *Server) Put(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	buf, _ := ioutil.ReadAll(r.Body)
	server.process(w, CmdPut, []byte(key), buf)
}

// Delete delete a key
func (server *Server) Delete(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	server.process(w, CmdDelete, []byte(key), nil)
}

func (server *Server) getByReadIndex(w http.ResponseWriter, key string) {
	future := server.raftServer.ReadIndex(DefaultClusterID)
	respCh, errCh := future.AsyncResponse()
	select {
	case resp := <-respCh:
		if resp != nil {
			klog.Errorf("process get %s failed: unexpected resp type: %T", key, resp)
			return
		}
		value, err := server.db.Get([]byte(key))
		if err != nil {
			klog.Errorf(fmt.Sprintf("[getByReadIndex]get %s with err %v", key, err))
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.Write(value)
		}
	case err := <-errCh:
		klog.Errorf("process get %s failed: %v", key, err)
		if err == raft.ErrNotLeader {
			server.replyNotLeader(w)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case <-time.After(DefaultRequestTimeout):
		klog.Errorf("process get %s timeout", key)
		w.WriteHeader(http.StatusRequestTimeout)
	}
}

func (server *Server) replyNotLeader(w http.ResponseWriter) {
	leader, term := server.raftServer.LeaderTerm(DefaultClusterID)
	w.Header().Add("leader", fmt.Sprintf("%d", leader))
	w.Header().Add("term", fmt.Sprintf("%d", term))
	node := server.cfg.FindClusterNode(leader)
	if node != nil {
		w.Header().Add("leader-host", node.Host)
		w.Header().Add("leader-addr", fmt.Sprintf("%s:%d", node.Host, node.HTTPPort))
	} else {
		w.Header().Add("leader-host", "")
		w.Header().Add("leader-addr", "")
	}
	w.WriteHeader(http.StatusMovedPermanently)
}

func (server *Server) process(w http.ResponseWriter, op CmdType, key, value []byte) {
	cmd := &Command{
		OP:    op,
		Key:   key,
		Value: value,
	}
	klog.Infof("start process command: %v", cmd)

	if server.leader != server.nodeID {
		server.replyNotLeader(w)
		return
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		klog.Errorf("marshal raft command failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 提交 cmd 到 raft propose channel
	future := server.raftServer.Submit(DefaultClusterID, data)
	respCh, errCh := future.AsyncResponse()
	select {
	case resp := <-respCh:
		switch r := resp.(type) {
		case []byte:
			w.Write(r)
		case nil:
		default:
			klog.Errorf("unknown resp type: %T", r)
			w.WriteHeader(http.StatusInternalServerError)
		}
	case err := <-errCh:
		klog.Errorf("process %v failed: %v", cmd.String(), err)
		if err != nil {
			klog.Infof("process %v not found", cmd.String())
			w.WriteHeader(http.StatusNotFound)
		} else if err == raft.ErrNotLeader {
			server.replyNotLeader(w)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case <-time.After(DefaultRequestTimeout):
		klog.Errorf("process %v timeout", cmd.String())
		w.WriteHeader(http.StatusRequestTimeout)
	}
}

// Apply implement raft StateMachine Apply method
func (server *Server) Apply(command []byte, index uint64) (interface{}, error) {
	klog.Infof("apply command at index(%v): %v", index, string(command))

	cmd := new(Command)
	err := json.Unmarshal(command, cmd)
	if err != nil {
		return nil, fmt.Errorf("unmarshal command failed: %v", command)
	}
	switch cmd.OP {
	case CmdQuorumGet:
		return server.db.Get(cmd.Key)
	case CmdPut:
		return nil, server.db.Put(cmd.Key, cmd.Value)
	case CmdDelete:
		return nil, server.db.Delete(cmd.Key)
	default:
		return nil, fmt.Errorf("invalid cmd type: %v", cmd.OP)
	}
}

// ApplyMemberChange implement raft.StateMachine
func (server *Server) ApplyMemberChange(confChange *proto.ConfChange, index uint64) (interface{}, error) {
	return nil, errors.New("not supported")
}

// Snapshot implement raft.StateMachine
func (server *Server) Snapshot() (proto.Snapshot, error) {
	return nil, errors.New("not supported")
}

// ApplySnapshot implement raft.StateMachine
func (server *Server) ApplySnapshot(peers []proto.Peer, iter proto.SnapIterator) error {
	return errors.New("not supported")
}

// HandleFatalEvent implement raft.StateMachine
func (server *Server) HandleFatalEvent(err *raft.FatalError) {
	klog.Fatalf("raft fatal error: %v", err)
}

// HandleLeaderChange implement raft.StateMachine
func (server *Server) HandleLeaderChange(leader uint64) {
	klog.Infof("raft leader change to %v", leader)
	server.leader = leader
}

func (server *Server) startRaft() {
	// start raft server
	serverConfig := raft.DefaultConfig()
	serverConfig.NodeID = server.nodeID
	serverConfig.Resolver = newCluster(server.cfg)
	serverConfig.TickInterval = time.Millisecond * 500
	serverConfig.ReplicateAddr = fmt.Sprintf(":%d", server.node.ReplicatePort)
	serverConfig.HeartbeatAddr = fmt.Sprintf(":%d", server.node.HeartbeatPort)
	raftServer, err := raft.NewRaftServer(serverConfig)
	if err != nil {
		klog.Fatalf("start raft server failed: %v", err)
	}
	server.raftServer = raftServer
	klog.Infof("raft server started.")

	// create raft
	walPath := path.Join(server.cfg.ServerCfg.DataPath, "wal")
	// raftStore := storage.NewMemoryStorage(raftServer, 1, 8192)
	raftStore, err := wal.NewStorage(walPath, &wal.Config{})
	if err != nil {
		klog.Fatalf("init raft log storage failed: %v", err)
	}
	rc := &raft.RaftConfig{
		ID: DefaultClusterID,
		// TODO 暂时使用 github.com/tiglabs/raft/storage/wal ，后续写个自己的wal库
		Storage:      raftStore, // 后端存储，可以是boltdb, leveldb or memory
		StateMachine: server,    // server 就是 state machine
	}
	for _, node := range server.cfg.ClusterCfg.Nodes {
		rc.Peers = append(rc.Peers, proto.Peer{
			Type:   proto.PeerNormal,
			ID:     node.NodeID,
			PeerID: node.NodeID,
		})
	}

	err = server.raftServer.CreateRaft(rc)
	if err != nil {
		klog.Fatalf("create raft failed: %v", err)
	}
	klog.Info("raft created.")
}

func (server *Server) initBoltDB() {
	server.db = newStore(path.Join(server.cfg.ServerCfg.DataPath, "my.db"))
}

// Run run server
func (server *Server) Run() {
	// init store
	server.initBoltDB()
	defer server.db.Close()

	// start raft
	server.startRaft()
	// start http server, block
	server.startHTTPServer()
}

// NewServer create kvs
func NewServer(nodeID uint64, cfg *Config) *Server {
	server := &Server{
		nodeID: nodeID,
		cfg:    cfg,
	}
	node := cfg.FindClusterNode(nodeID)
	if node == nil {
		klog.Fatalf("could not find self node(%v) in cluster config: (%v)", nodeID, cfg.String())
	}
	server.node = node
	return server
}
