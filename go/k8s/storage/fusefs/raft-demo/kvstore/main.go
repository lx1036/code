package main

import (
	"flag"

	"k8s.io/klog/v2"
)

// INFO: https://github.com/tiglabs/raft/blob/master/test/kvs/main.go

var nodeID = flag.Uint64("node", 0, "current node id")
var confFile = flag.String("conf", "", "config file path")

// debug:
// rm -rf ./raft
// go run . --node=1 --conf=./conf/kvstore.toml
// go run . --node=2 --conf=./conf/kvstore.toml
// go run . --node=3 --conf=./conf/kvstore.toml

// rm -rf ./raft
// go run . --node=1 --conf=./kvstore.toml

// curl -X PUT -d '{"hello":"world"}' localhost:7771/kvs/hello
// curl -X GET localhost:7771/kvs/hello
// curl -X GET "localhost:7771/kvs/hello?level=log"
func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	// load config
	cfg := LoadConfig(*confFile, *nodeID)

	// start server
	server := NewServer(*nodeID, cfg)
	server.Run()
}
