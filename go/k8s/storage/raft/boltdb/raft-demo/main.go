// INFO: 参考文档 https://github.com/vision9527/raft-demo/blob/election-1/README.md
//  https://github.com/talkgo/night/blob/master/content/night/104-2020-09-13-hashicorp-raft.md
//  https://talkgo.org/t/topic/882

package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s-lx1036/k8s/storage/raft/boltdb"

	"github.com/hashicorp/raft"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/klog/v2"
)

var (
	httpAddr    string
	raftAddr    string
	raftId      string
	raftCluster string
)

var (
	isLeader bool
)

func init() {
	flag.StringVar(&httpAddr, "http_addr", "127.0.0.1:7001", "http listen addr")
	flag.StringVar(&raftAddr, "raft_addr", "127.0.0.1:7000", "raft listen addr")
	flag.StringVar(&raftId, "raft_id", "1", "raft id")
	flag.StringVar(&raftCluster, "raft_cluster", "1/127.0.0.1:7000,2/127.0.0.1:8000,3/127.0.0.1:9000", "cluster info")
}

type HttpServer struct {
	raft *raft.Raft
	fsm  *Fsm
}

func (h *HttpServer) Set(w http.ResponseWriter, r *http.Request) {
	if !isLeader {
		fmt.Fprintf(w, "not leader")
		return
	}
	vars := r.URL.Query()
	key := vars.Get("key")
	value := vars.Get("value")
	if key == "" || value == "" {
		fmt.Fprintf(w, "error key or value")
		return
	}

	data := "set" + "," + key + "," + value
	future := h.raft.Apply([]byte(data), 5*time.Second)
	if err := future.Error(); err != nil {
		fmt.Fprintf(w, "error:"+err.Error())
		return
	}
	fmt.Fprintf(w, "ok")
	return
}

func (h *HttpServer) Get(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	key := vars.Get("key")
	if key == "" {
		fmt.Fprintf(w, "error key")
		return
	}

	value := h.fsm.Data[key]
	fmt.Fprintf(w, value)
	return
}

// go run . --http_addr=127.0.0.1:7001 --raft_addr=127.0.0.1:7000 --raft_id=1 --raft_cluster=1/127.0.0.1:7000,2/127.0.0.1:8000,3/127.0.0.1:9000
// go run . --http_addr=127.0.0.1:8001 --raft_addr=127.0.0.1:8000 --raft_id=2 --raft_cluster=1/127.0.0.1:7000,2/127.0.0.1:8000,3/127.0.0.1:9000
// go run . --http_addr=127.0.0.1:9001 --raft_addr=127.0.0.1:9000 --raft_id=3 --raft_cluster=1/127.0.0.1:7000,2/127.0.0.1:8000,3/127.0.0.1:9000

// curl http://127.0.0.1:7001/set?key=hello&value=world
// curl http://127.0.0.1:7001/get?key=hello
func main() {
	flag.Parse()

	raftDir := "raft/raft_" + raftId
	os.MkdirAll(raftDir, 0700)

	// INFO: (1)初始化 raft 对象
	rf, fsm, err := NewRaft(raftAddr, raftId, raftDir)
	if err != nil {
		klog.Fatal(err)
	}

	// INFO: (2)start raft cluster
	servers := rf.GetConfiguration().Configuration().Servers
	if len(servers) > 0 {
		return
	}
	peerArray := strings.Split(raftCluster, ",")
	if len(peerArray) == 0 {
		return
	}
	var configuration raft.Configuration
	for _, peerInfo := range peerArray {
		peer := strings.Split(peerInfo, "/")
		id := peer[0]
		addr := peer[1]
		server := raft.Server{
			ID:      raft.ServerID(id),
			Address: raft.ServerAddress(addr),
		}
		configuration.Servers = append(configuration.Servers, server)
	}
	rf.BootstrapCluster(configuration)

	// 监听leader变化
	go func() {
		for leader := range rf.LeaderCh() {
			isLeader = leader
		}
	}()

	// 启动http server
	httpServer := &HttpServer{
		raft: rf,
		fsm:  fsm,
	}

	http.HandleFunc("/set", httpServer.Set)
	http.HandleFunc("/get", httpServer.Get)
	go http.ListenAndServe(httpAddr, nil)

	// stop channel closed on SIGTERM and SIGINT
	stopCh := genericapiserver.SetupSignalHandler()
	<-stopCh

	err = rf.Shutdown().Error()
	if err != nil {
		klog.Error(err)
	}
}

func NewRaft(raftAddr, raftId, raftDir string) (*raft.Raft, *Fsm, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(raftId)
	// config.HeartbeatTimeout = 1000 * time.Millisecond
	// config.ElectionTimeout = 1000 * time.Millisecond
	// config.CommitTimeout = 1000 * time.Millisecond
	addr, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		return nil, nil, err
	}
	transport, err := raft.NewTCPTransport(raftAddr, addr, 2, 5*time.Second, os.Stderr)
	if err != nil {
		return nil, nil, err
	}
	snapshots, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		return nil, nil, err
	}

	logStore, err := boltdb.NewBoltStore(filepath.Join(raftDir, "raft-log.db"))
	if err != nil {
		return nil, nil, err
	}
	stableStore, err := boltdb.NewBoltStore(filepath.Join(raftDir, "raft-stable.db"))
	if err != nil {
		return nil, nil, err
	}

	fsm := new(Fsm)
	fsm.Data = map[string]string{}
	rf, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, nil, err
	}

	return rf, fsm, nil
}
