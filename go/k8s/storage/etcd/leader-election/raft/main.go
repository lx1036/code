package main

import (
	"flag"
	"strings"

	"k8s.io/klog/v2"

	raft "k8s-lx1036/k8s/storage/etcd/leader-election/raft/pkg"
)

var (
	cluster = flag.String("cluster", "http://127.0.0.1:12379", "comma separated cluster peers")
	id      = flag.Int("id", 1, "node id")
	port    = flag.Int("port", 12379, "key-value server port")
	join    = flag.Bool("join", false, "join an existing cluster")
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	proposeC := make(chan string)
	defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)
	var kvStore *raft.KVStore
	getSnapshot := func() ([]byte, error) { return kvStore.GetSnapshot() }
	commitC, errorC, snapshotterReady := raft.NewRaftNode(*id, strings.Split(*cluster, ","), *join, getSnapshot, proposeC, confChangeC)

	kvStore = raft.NewKVStore(<-snapshotterReady, proposeC, commitC, errorC)

	// the key-value http handler will propose updates to raft
	raft.ServeHttpKVAPI(kvStore, *port, confChangeC, errorC)
}
