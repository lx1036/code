package meta

import (
	"flag"
	"k8s.io/klog/v2"
	"os"
	"testing"
)

var (
	MasterAddr = flag.String("masterAddr", "", "")
	owner      = flag.String("owner", "", "")
	volName    = flag.String("vol", "", "")
)

func TestMeta(test *testing.T) {
	flag.Parse()
	if len(*MasterAddr) == 0 || len(*owner) == 0 || len(*volName) == 0 {
		klog.Error("--masterAddr or --owner or --vol is needed")
		os.Exit(1)
	}
	//volName := "pvc-liuxiang"
	meta, err := NewMetaWrapper(*volName, *owner, *MasterAddr)
	if err != nil {
		klog.Fatal(err)
	}

	parentID := uint64(0)
	dentries, err := meta.ReadDir_ll(parentID)
	if err != nil {
		klog.Fatal(err)
	}
	for _, dentry := range dentries {
		klog.Info(dentry)
	}
}
