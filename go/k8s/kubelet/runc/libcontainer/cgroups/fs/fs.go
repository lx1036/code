package fs

import (
	"sync"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups"
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

var (
	HugePageSizes, _ = cgroups.GetHugePageSize()

	subsystemsLegacy = subsystemSet{
		&CpusetGroup{},
		&DevicesGroup{},
		&MemoryGroup{},
		&CpuGroup{},
		&CpuacctGroup{},
		&PidsGroup{},
		&BlkioGroup{},
		&HugetlbGroup{},
		&NetClsGroup{},
		&NetPrioGroup{},
		&PerfEventGroup{},
		&FreezerGroup{},
		&NameGroup{GroupName: "name=systemd", Join: true},
	}
)

type cgroupData struct {
	root      string
	innerPath string
	config    *configs.Cgroup
	pid       int
}

type subsystemSet []subsystem
type subsystem interface {
	// Name returns the name of the subsystem.
	Name() string
	// Returns the stats, as 'stats', corresponding to the cgroup under 'path'.
	GetStats(path string, stats *cgroups.Stats) error
	// Removes the cgroup represented by 'cgroupData'.
	Remove(*cgroupData) error
	// Creates and joins the cgroup represented by 'cgroupData'.
	Apply(*cgroupData) error
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
