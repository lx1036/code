package main

import (
	"flag"
	"path/filepath"
	"testing"

	"k8s.io/klog/v2"
)

func TestClean(test *testing.T) {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	name := filepath.Clean("/tmp/csi_example4/plugin.sock")
	klog.Infof("name: %s", name)
}
