package main

import (
	"flag"

	"k8s-lx1036/k8s/storage/csi/csi-drivers/pkg/hostpath"

	"k8s.io/klog/v2"
)

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	driverName = flag.String("drivername", "csi-hostpath", "name of the driver")
	nodeID     = flag.String("nodeid", "", "node id")
)

//debug: go run . --endpoint tcp://127.0.0.1:10000 --nodeid minikube
func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	driver := hostpath.GetHostPathDriver()
	driver.Run(*driverName, *nodeID, *endpoint)
}
