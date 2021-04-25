package main

import (
	"testing"

	"github.com/google/cadvisor/fs"
	//"github.com/stretchr/testify/assert"
	//"k8s.io/utils/mount"
)

func TestMountInfoFromDir(test *testing.T) {
	//as := assert.New(test)

	fsInfo, err := fs.NewFsInfo(fs.Context{})
	if err != nil {
		panic(err)
	}
	testDirs := []string{"/var/lib/kubelet", "/var/lib/rancher"}
	for _, testDir := range testDirs {
		_, _ = fsInfo.GetDirFsDevice(testDir)
	}
}
