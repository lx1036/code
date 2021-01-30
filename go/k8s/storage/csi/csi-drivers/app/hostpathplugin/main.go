package main

import (
	"flag"
	"k8s-lx1036/k8s/storage/csi/csi-drivers/pkg/hostpath"
	"os"
)

func init() {
	flag.Set("logtostderr", "true")
}

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	driverName = flag.String("drivername", "csi-hostpath", "name of the driver")
	nodeID     = flag.String("nodeid", "", "node id")
)

//debug: go run . --endpoint tcp://127.0.0.1:10000 --nodeid CSINode -v=5
func main() {
	flag.Parse()

	handle()
	os.Exit(0)
}

func handle() {
	driver := hostpath.GetHostPathDriver()
	driver.Run(*driverName, *nodeID, *endpoint)
}
