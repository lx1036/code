package main

import (
	"flag"

	"k8s.io/klog/v2"
)

var nodeID = flag.Uint64("node", 0, "current node id")
var confFile = flag.String("conf", "", "config file path")

// debug:
// go run . --node=1 --conf=./conf/kvstore.toml
// go run . --node=2 --conf=./conf/kvstore.toml
// go run . --node=3 --conf=./conf/kvstore.toml
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
