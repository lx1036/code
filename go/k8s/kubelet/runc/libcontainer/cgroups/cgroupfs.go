package cgroups

import (
	"sync"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

// INFO: cgroupfs-driver cgroup v1

// INFO: 本包是 runc cgroups client-go，同时也已经实现了自己的 containerd cgroups client-go

const defaultDirPerm = 0755

type Manager interface {
	// Applies cgroup configuration to the process with the specified pid
	Apply(pid int) error

	// Returns the PIDs inside the cgroup set
	GetPids() ([]int, error)

	// Returns the PIDs inside the cgroup set & all sub-cgroups
	GetAllPids() ([]int, error)

	// Returns statistics for the cgroup set
	GetStats() (*Stats, error)

	// Toggles the freezer cgroup according with specified state
	Freeze(state configs.FreezerState) error

	// Destroys the cgroup set
	Destroy() error

	// Path returns a cgroup path to the specified controller/subsystem.
	// For cgroupv2, the argument is unused and can be empty.
	Path(string) string

	// Sets the cgroup as configured.
	Set(container *configs.Config) error

	// GetPaths returns cgroup path(s) to save in a state file in order to restore later.
	//
	// For cgroup v1, a key is cgroup subsystem name, and the value is the path
	// to the cgroup for this subsystem.
	//
	// For cgroup v2 unified hierarchy, a key is "", and the value is the unified path.
	GetPaths() map[string]string

	// GetCgroups returns the cgroup data as configured.
	GetCgroups() (*configs.Cgroup, error)

	// GetFreezerState retrieves the current FreezerState of the cgroup.
	GetFreezerState() (configs.FreezerState, error)

	// Whether the cgroup path exists or not
	Exists() bool
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

	data, err := getCgroupData(m.cgroups, pid)
	if err != nil {
		return err
	}

	m.paths = make(map[string]string)

	for _, sys := range m.getSubsystems() {
		p, err := data.path(sys.Name())
		if err != nil {
			return err
		}

		m.paths[sys.Name()] = p

		if err := sys.Apply(data); err != nil {
			return err
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

func (m *manager) GetStats() (*Stats, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := NewStats()
	for name, path := range m.paths {
		sys, err := m.getSubsystems().Get(name)
		if err != nil || err == errSubsystemDoesNotExist || !PathExists(path) {
			continue
		}
		if err := sys.GetStats(path, stats); err != nil {
			return nil, err
		}
	}

	return stats, nil
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

func NewManager(cg *configs.Cgroup, paths map[string]string, rootless bool) Manager {
	return &manager{
		cgroups:  cg,
		paths:    paths,
		rootless: rootless,
	}
}
