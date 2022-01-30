package cgroups

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"

	"k8s.io/klog/v2"
)

type cgroupTestUtil struct {
	// cgroup data to use in tests.
	CgroupData *cgroupData

	// Path to the mock cgroup directory.
	CgroupPath string

	// Temporary directory to store mock cgroup filesystem.
	tempDir string
	t       *testing.T
}

// Creates a new test util for the specified subsystem
func NewCgroupTestUtil(subsystem string, t *testing.T) *cgroupTestUtil {
	d := &cgroupData{
		config: &configs.Cgroup{},
	}
	d.config.Resources = &configs.Resources{}
	tempDir, err := ioutil.TempDir("", "cgroup_test")
	if err != nil {
		t.Fatal(err)
	}

	d.root = tempDir
	testCgroupPath := filepath.Join(d.root, subsystem)

	klog.Infof("testCgroupPath: %s", testCgroupPath)

	// Ensure the full mock cgroup path exists.
	err = os.MkdirAll(testCgroupPath, 0755)
	if err != nil {
		t.Fatal(err)
	}
	return &cgroupTestUtil{CgroupData: d, CgroupPath: testCgroupPath, tempDir: tempDir, t: t}
}

func (c *cgroupTestUtil) cleanup() {
	os.RemoveAll(c.tempDir)
}

// Write the specified contents on the mock of the specified cgroup files.
func (c *cgroupTestUtil) writeFileContents(fileContents map[string]string) {
	for file, contents := range fileContents {
		err := WriteFile(c.CgroupPath, file, contents)
		if err != nil {
			c.t.Fatal(err)
		}
	}
}

func TestGetCgroupMounts(t *testing.T) {
	mounts, err := getCgroupMountsV1(true)
	if err != nil {
		panic(err)
	}

	for _, mount := range mounts {
		klog.Infof("Mountpoint %s, Root %s, Subsystems %v", mount.Mountpoint, mount.Root, mount.Subsystems)
	}
	// output:
	// Mountpoint /sys/fs/cgroup/systemd, Root /, Subsystems [systemd]
	// Mountpoint /sys/fs/cgroup/cpuset, Root /, Subsystems [cpuset]
	// Mountpoint /sys/fs/cgroup/cpu,cpuacct, Root /, Subsystems [cpuacct cpu]
	// Mountpoint /sys/fs/cgroup/memory, Root /, Subsystems [memory]
	// Mountpoint /sys/fs/cgroup/devices, Root /, Subsystems [devices]
	// Mountpoint /sys/fs/cgroup/freezer, Root /, Subsystems [freezer]
	// Mountpoint /sys/fs/cgroup/net_cls, Root /, Subsystems [net_cls]
	// Mountpoint /sys/fs/cgroup/blkio, Root /, Subsystems [blkio]
	// Mountpoint /sys/fs/cgroup/perf_event, Root /, Subsystems [perf_event]
	// Mountpoint /sys/fs/cgroup/hugetlb, Root /, Subsystems [hugetlb]

}
