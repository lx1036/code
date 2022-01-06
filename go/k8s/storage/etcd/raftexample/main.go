package main

import (
	"flag"
	"strings"

	raft "k8s-lx1036/k8s/storage/etcd/raftexample/pkg"

	"go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
)

var (
	cluster = flag.String("cluster", "http://127.0.0.1:12379", "comma separated cluster peers")
	id      = flag.Int("id", 1, "node id")
	port    = flag.Int("port", 12380, "key-value server port")
	join    = flag.Bool("join", false, "join an existing cluster")
)

// go run . --id 1 --cluster http://127.0.0.1:12379 --port 12380

// go run . --id 1 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 12380
// go run . --id 2 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 22380
// go run . --id 3 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 32380
func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	proposeC := make(chan string)
	defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	var kvStore *raft.KVStore
	getSnapshotDataFromStore := func() ([]byte, error) {
		return kvStore.GetSnapshot()
	}
	commitC, errorC, snapshotterReady := raft.NewRaftNode(*id, strings.Split(*cluster, ","), *join, getSnapshotDataFromStore, proposeC, confChangeC)

	kvStore = raft.NewKVStore(<-snapshotterReady, proposeC, commitC, errorC)

	// the key-value http handler will propose updates to raft
	raft.ServeHttpKVAPI(kvStore, *port, confChangeC, errorC)
}
