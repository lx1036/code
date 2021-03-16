package main

import (
	"testing"

	"github.com/google/cadvisor/fs"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/mount"
)

func TestMountInfoFromDir(test *testing.T) {
	as := assert.New(test)
	fsInfo := &fs.RealFsInfo{
		mounts: map[string]mount.Info{
			"/": {},
		},
	}
	testDirs := []string{"/var/lib/kubelet", "/var/lib/rancher"}
	for _, testDir := range testDirs {
		_, found := fsInfo.mountInfoFromDir(testDir)
		as.True(found, "failed to find MountInfo %s from FsInfo %s", testDir, fsInfo)
	}
}
