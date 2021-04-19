package fs

import (
	"errors"
	"path/filepath"
	"sync"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups"
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

// The absolute path to the root of the cgroup hierarchies.
var cgroupRootLock sync.Mutex
var cgroupRoot string

type cgroupData struct {
	root      string
	innerPath string
	config    *configs.Cgroup
	pid       int
}

func (raw *cgroupData) path(subsystem string) (string, error) {
	// If the cgroup name/path is absolute do not look relative to the cgroup of the init process.
	if filepath.IsAbs(raw.innerPath) {
		mnt, err := cgroups.FindCgroupMountpoint(raw.root, subsystem)
		// If we didn't mount the subsystem, there is no point we make the path.
		if err != nil {
			return "", err
		}

		// Sometimes subsystems can be mounted together as 'cpu,cpuacct'.
		return filepath.Join(raw.root, filepath.Base(mnt), raw.innerPath), nil
	}

	// Use GetOwnCgroupPath instead of GetInitCgroupPath, because the creating
	// process could in container and shared pid namespace with host, and
	// /proc/1/cgroup could point to whole other world of cgroups.
	parentPath, err := cgroups.GetOwnCgroupPath(subsystem)
	if err != nil {
		return "", err
	}

	return filepath.Join(parentPath, raw.innerPath), nil
}

const defaultCgroupRoot = "../mock/sys/fs/cgroup"

func tryDefaultCgroupRoot() string {
	path, err := filepath.Abs(defaultCgroupRoot)
	if err != nil {
		panic(err)
	}

	return path
}

// Gets the cgroupRoot.
func getCgroupRoot() (string, error) {
	cgroupRootLock.Lock()
	defer cgroupRootLock.Unlock()

	if cgroupRoot != "" {
		return cgroupRoot, nil
	}

	// fast path
	cgroupRoot = tryDefaultCgroupRoot()
	if cgroupRoot != "" {
		return cgroupRoot, nil
	}

	return "", nil
}

func getCgroupData(c *configs.Cgroup, pid int) (*cgroupData, error) {
	root, err := getCgroupRoot()
	if err != nil {
		return nil, err
	}

	if (c.Name != "" || c.Parent != "") && c.Path != "" {
		return nil, errors.New("cgroup: either Path or Name and Parent should be used")
	}

	cgPath := filepath.Clean(c.Path)
	cgParent := filepath.Clean(c.Parent)
	cgName := filepath.Clean(c.Name)
	innerPath := cgPath
	if innerPath == "" {
		innerPath = filepath.Join(cgParent, cgName)
	}

	return &cgroupData{
		root:      root,
		innerPath: innerPath,
		config:    c,
		pid:       pid,
	}, nil
}
