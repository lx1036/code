package fakesysfs

import (
	"os"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils/sysfs"
)

// If we extend sysfs to support more interfaces, it might be worth making this a mock instead of a fake.
type FileInfo struct {
	EntryName string
}

type FakeSysFs struct {
	info  FileInfo
	cache sysfs.CacheInfo

	nodesPaths  []string
	nodePathErr error

	cpusPaths  map[string][]string
	cpuPathErr error

	coreThread map[string]string
	coreIDErr  map[string]error

	physicalPackageIDs   map[string]string
	physicalPackageIDErr map[string]error

	memTotal string
	memErr   error

	hugePages    []os.FileInfo
	hugePagesErr error

	hugePagesNr    map[string]string
	hugePagesNrErr error

	onlineCPUs map[string]interface{}
}
