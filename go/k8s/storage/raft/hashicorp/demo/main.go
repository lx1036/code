package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	boltdb "k8s-lx1036/k8s/storage/raft/hashicorp/bolt-store"

	"github.com/hashicorp/raft"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/klog/v2"
)

// INFO: 参考文档 https://github.com/vision9527/raft-demo/blob/election-1/README.md
//  https://github.com/talkgo/night/blob/master/content/night/104-2020-09-13-hashicorp-raft.md
//  https://talkgo.org/t/topic/882

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

	value := h.fsm.kvstore.Get(key)
	fmt.Fprintf(w, value)
	return
}

// INFO: https://github.com/vision9527/raft-demo/blob/master/README.md
//  (1) leader election

// go run . --http_addr=127.0.0.1:7001 --raft_addr=127.0.0.1:7000 --raft_id=1 --raft_cluster=1/127.0.0.1:7000,2/127.0.0.1:8000,3/127.0.0.1:9000
// go run . --http_addr=127.0.0.1:8001 --raft_addr=127.0.0.1:8000 --raft_id=2 --raft_cluster=1/127.0.0.1:7000,2/127.0.0.1:8000,3/127.0.0.1:9000
// go run . --http_addr=127.0.0.1:9001 --raft_addr=127.0.0.1:9000 --raft_id=3 --raft_cluster=1/127.0.0.1:7000,2/127.0.0.1:8000,3/127.0.0.1:9000

// curl "http://127.0.0.1:7001/set?key=hello&value=world" # http://127.0.0.1:7001 是 leader
// curl "http://127.0.0.1:7001/get?key=hello"
// curl "http://127.0.0.1:8001/get?key=hello"
// curl "http://127.0.0.1:8001/set?key=hello&value=world"
// 断掉之后重新拉起，记得清空数据 `rm -rf ./raft/raft_3`
func main() {
	flag.Parse()

	raftDir := "raft/raft_" + raftId
	raftDir = filepath.Clean(raftDir)
	if _, err := os.Stat(raftDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(raftDir, 0700)
		} else {
			klog.Fatalf(fmt.Sprintf("%s is err:%v", raftDir, err))
		}
	}

	// INFO: (1)初始化 raft 对象
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(raftId)
	config.HeartbeatTimeout = 1000 * time.Millisecond
	config.ElectionTimeout = config.HeartbeatTimeout * 10 // electionTimeout=heartbeatTimeout * 10
	config.BatchApplyCh = true
	//config.Logger = NewLogger()
	config.SnapshotThreshold = 5               // 有 5 个 log entry 就可以触发 snapshot
	config.SnapshotInterval = time.Second * 60 // 每 [60s, 120) 检查是否达到 snapshot threshold
	config.TrailingLogs = 3                    // compact logs 之后，距离 snapshotIndex，还可以有 TrailingLogs 个 logs
	addr, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		klog.Fatal(err)
	}
	transport, err := raft.NewTCPTransport(raftAddr, addr, 2, 5*time.Second, os.Stderr)
	if err != nil {
		klog.Fatal(err)
	}
	snapshots, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		klog.Fatal(err)
	}
	store, err := boltdb.NewBoltStore(filepath.Join(raftDir, "raft-log.db"))
	if err != nil {
		klog.Fatal(err)
	}
	fsm := &Fsm{
		kvstore: NewKVStore(),
	}
	r, err := raft.NewRaft(config, fsm, store, store, snapshots, transport)
	if err != nil {
		klog.Fatal(err)
	}

	// INFO: (2)start raft cluster
	var configuration raft.Configuration
	servers := r.GetConfiguration().Configuration().Servers
	if len(servers) > 0 {
		configuration.Servers = servers
	} else {
		peerArray := strings.Split(raftCluster, ",")
		if len(peerArray) == 0 {
			return
		}
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
	}
	r.BootstrapCluster(configuration)
	klog.Infof(fmt.Sprintf("raft is started"))

	// 监听leader变化
	go func() {
		for leader := range r.LeaderCh() {
			isLeader = leader
		}
	}()

	// 启动http server
	httpServer := &HttpServer{
		raft: r,
		fsm:  fsm,
	}

	http.HandleFunc("/set", httpServer.Set)
	http.HandleFunc("/get", httpServer.Get)
	go http.ListenAndServe(httpAddr, nil)

	dumpData()

	// stop channel closed on SIGTERM and SIGINT
	stopCh := genericapiserver.SetupSignalHandler()
	<-stopCh

	err = r.Shutdown().Error()
	if err != nil {
		klog.Error(err)
	}
}

func dumpData() {
	go func() {
		time.Sleep(time.Second * 30)
		if isLeader {
			for i := 1; i <= 10; i++ {
				resp, err := http.Get(fmt.Sprintf("http://%s/set?key=hello%d&value=world%d", httpAddr, i, i))
				if err != nil {
					klog.Error(err)
					return
				}
				data, _ := io.ReadAll(resp.Body)
				klog.Info(string(data))
			}
		}
	}()
}
