package main

import (
	"flag"
	"fmt"
	"k8s-lx1036/k8s/storage/raft/wal"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"strings"
)

var (
	// 1
	id    = flag.Int("id", -1, "node id")
	nodes = flag.String("nodes", "", "all nodes")
)

// go run . --id=1 --nodes=127.0.0.1:8080,127.0.0.1:8081,127.0.0.1:8082
// go run . --id=2 --nodes=127.0.0.1:8080,127.0.0.1:8081,127.0.0.1:8082
// go run . --id=3 --nodes=127.0.0.1:8080,127.0.0.1:8081,127.0.0.1:8082
func main() {
	flag.Parse()
	var cluster []*wal.Node
	var currentNode *wal.Node
	nodeSlice := strings.Split(*nodes, ",")
	for key, value := range nodeSlice {
		addrPort := strings.Split(value, ":")
		port, _ := strconv.Atoi(addrPort[1])
		if key+1 == *id {
			currentNode = &wal.Node{
				ID:   *id,
				Addr: addrPort[0],
				Port: port,
			}
		}
		cluster = append(cluster, &wal.Node{
			ID:   *id,
			Addr: addrPort[0],
			Port: port,
		})
	}

	stopCh := genericapiserver.SetupSignalHandler()

	config := &wal.Config{
		Node:          currentNode,
		Nodes:         cluster,
		HeartBeatTime: 1000,
	}
	server, err := NewServer(config)
	if err != nil {
		klog.Fatal(err)
	}
	err = server.Start()
	if err != nil {
		klog.Fatal(err)
	}

	<-stopCh
	klog.Info("shutdown raft server")
}

type Server struct {
	server *wal.Server

	stateMachine *StateMachine

	addr string
}

func (server *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value, ok := server.stateMachine.data[key]
	if !ok {
		errMsg := fmt.Sprintf("no key %s in state machine", key)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}
	w.Write([]byte(value))
}

func LeaderCheck(handler func(http.ResponseWriter, *http.Request), server *wal.Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if server.Role() == wal.Leader {
			handler(w, r)
		} else {
			leader, err := server.Leader()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			http.Redirect(w, r, fmt.Sprintf("http://%s:%d", leader.Addr, leader.Port), 302)
		}
	}
}

func (server *Server) Start() error {
	err := server.server.Start()
	if err != nil {
		return err
	}

	http.HandleFunc("/put", LeaderCheck(server.handlePut, server.server))
	// http.HandleFunc("/get", LeaderCheck(server.handleGet, server.server))
	http.HandleFunc("/get", server.handleGet)
	//http.HandleFunc("/state", server.handleState)
	go func() {
		http.ListenAndServe(server.addr, nil) // 127.0.0.1:8080
	}()

	return nil
}

func NewServer(config *wal.Config) (*Server, error) {
	stateMachine := &StateMachine{
		data: map[string]string{},
	}
	server, err := wal.NewServer(config, stateMachine, config.Nodes, config.Node)
	if err != nil {
		return nil, err
	}

	s := &Server{
		server:       server,
		stateMachine: stateMachine,
		addr:         fmt.Sprintf("%s:%d", config.Node.Addr, config.Node.Port),
	}

	return s, nil
}

type StateMachine struct {
	data map[string]string
}

func (stateMachine *StateMachine) Apply(data []byte) error {
	kv := strings.Split(string(data), "=")
	stateMachine.data[kv[0]] = kv[1]
	return nil
}
