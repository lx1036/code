package fs

import (
	"errors"
	"path/filepath"
	"sync"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups"
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

var (
	HugePageSizes, _ = cgroups.GetHugePageSize()

	subsystemsLegacy = subsystemSet{
		&CpusetGroup{},
		//&DevicesGroup{},
		//&MemoryGroup{},
		//&CpuGroup{},
		//&CpuacctGroup{},
		//&PidsGroup{},
		//&BlkioGroup{},
		//&HugetlbGroup{},
		//&NetClsGroup{},
		//&NetPrioGroup{},
		//&PerfEventGroup{},
		//&FreezerGroup{},
		//&NameGroup{GroupName: "name=systemd", Join: true},
	}
)

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

type subsystemSet []subsystem
type subsystem interface {
	// Name returns the name of the subsystem.
	Name() string
	// Returns the stats, as 'stats', corresponding to the cgroup under 'path'.
	GetStats(path string, stats *cgroups.Stats) error
	// Creates and joins the cgroup represented by 'cgroupData'.
	Apply(c *cgroupData) error
	// Set the cgroup represented by cgroup.
	Set(path string, cgroup *configs.Cgroup) error
}

type manager struct {
	mu       sync.Mutex
	cgroups  *configs.Cgroup
	rootless bool // ignore permission-related errors
	paths    map[string]string
}

func (m *manager) getSubsystems() subsystemSet {
	return subsystemsLegacy
}

// The absolute path to the root of the cgroup hierarchies.
var cgroupRootLock sync.Mutex
var cgroupRoot string

const defaultCgroupRoot = "./mock/sys/fs/cgroup"

func tryDefaultCgroupRoot() string {
	return defaultCgroupRoot
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

func (m *manager) Apply(pid int) error {
	if m.cgroups == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	var c = m.cgroups
	d, err := getCgroupData(m.cgroups, pid)
	if err != nil {
		return err
	}

	m.paths = make(map[string]string)
	if c.Paths != nil {

	}

	for _, sys := range m.getSubsystems() {
		p, err := d.path(sys.Name())
		if err != nil {

		}

		m.paths[sys.Name()] = p

		if err := sys.Apply(d); err != nil {

		}
	}

	return nil
}

func (m *manager) GetPids() ([]int, error) {
	panic("implement me")
}

func (m *manager) GetAllPids() ([]int, error) {
	panic("implement me")
}

func (m *manager) GetStats() (*cgroups.Stats, error) {
	panic("implement me")
}

func (m *manager) Freeze(state configs.FreezerState) error {
	panic("implement me")
}

func (m *manager) Destroy() error {
	panic("implement me")
}

func (m *manager) Path(s string) string {
	panic("implement me")
}

func (m *manager) Set(container *configs.Config) error {
	panic("implement me")
}

func (m *manager) GetPaths() map[string]string {
	panic("implement me")
}

func (m *manager) GetCgroups() (*configs.Cgroup, error) {
	panic("implement me")
}

func (m *manager) GetFreezerState() (configs.FreezerState, error) {
	panic("implement me")
}

func (m *manager) Exists() bool {
	panic("implement me")
}

func NewManager(cg *configs.Cgroup, paths map[string]string, rootless bool) cgroups.Manager {
	return &manager{
		cgroups:  cg,
		paths:    paths,
		rootless: rootless,
	}
}
