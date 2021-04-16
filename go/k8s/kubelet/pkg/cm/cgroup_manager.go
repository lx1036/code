package cm

// INFO: cgroup 客户端

import (
	"fmt"
	"path"
	"strings"

	libcontainercgroups "k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups"
	cgroupfs "k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups/fs"
	cgroupsystemd "k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups/systemd"
	libcontainerconfigs "k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

// MemoryStats holds the on-demand statistics from the memory cgroup
type MemoryStats struct {
	// Memory usage (in bytes).
	Usage int64
}

// ResourceStats holds on-demand statistics from various cgroup subsystems
type ResourceStats struct {
	// Memory statistics.
	MemoryStats *MemoryStats
}

// CgroupManager allows for cgroup management.
// Supports Cgroup Creation ,Deletion and Updates.
type CgroupManager interface {
	// Create creates and applies the cgroup configurations on the cgroup.
	// It just creates the leaf cgroups.
	// It expects the parent cgroup to already exist.
	Create(*CgroupConfig) error
	// Destroy the cgroup.
	Destroy(*CgroupConfig) error
	// Update cgroup configuration.
	Update(*CgroupConfig) error
	// Exists checks if the cgroup already exists
	Exists(name CgroupName) bool
	// Name returns the literal cgroupfs name on the host after any driver specific conversions.
	// We would expect systemd implementation to make appropriate name conversion.
	// For example, if we pass {"foo", "bar"}
	// then systemd should convert the name to something like
	// foo.slice/foo-bar.slice
	Name(name CgroupName) string
	// CgroupName converts the literal cgroupfs name on the host to an internal identifier.
	CgroupName(name string) CgroupName
	// Pids scans through all subsystems to find pids associated with specified cgroup.
	Pids(name CgroupName) []int
	// ReduceCPULimits reduces the CPU CFS values to the minimum amount of shares.
	ReduceCPULimits(cgroupName CgroupName) error
	// GetResourceStats returns statistics of the specified cgroup as read from the cgroup fs.
	GetResourceStats(name CgroupName) (*ResourceStats, error)
}

// ResourceConfig holds information about all the supported cgroup resource parameters.
type ResourceConfig struct {
	// Memory limit (in bytes).
	Memory *int64
	// CPU shares (relative weight vs. other containers).
	CpuShares *uint64
	// CPU hardcap limit (in usecs). Allowed cpu time in a given period.
	CpuQuota *int64
	// CPU quota period.
	CpuPeriod *uint64
	// HugePageLimit map from page size (in bytes) to limit (in bytes)
	HugePageLimit map[int64]int64
	// Maximum number of pids
	PidsLimit *int64
}

type CgroupConfig struct {
	// Fully qualified name prior to any driver specific conversions.
	Name CgroupName
	// ResourceParameters contains various cgroups settings to apply.
	ResourceParameters *ResourceConfig
}

// CgroupName is the abstract name of a cgroup prior to any driver specific conversion.
// It is specified as a list of strings from its individual components, such as:
// {"kubepods", "burstable", "pod1234-abcd-5678-efgh"}
type CgroupName []string

var RootCgroupName = CgroupName([]string{})

func (cgroupName CgroupName) ToCgroupfs() string {
	return "/" + path.Join(cgroupName...)
}

func escapeSystemdCgroupName(part string) string {
	return strings.Replace(part, "-", "_", -1)
}

func unescapeSystemdCgroupName(part string) string {
	return strings.Replace(part, "_", "-", -1)
}

// cgroupName.ToSystemd converts the internal cgroup name to a systemd name.
// For example, the name {"kubepods", "burstable", "pod1234-abcd-5678-efgh"} becomes
// "/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod1234_abcd_5678_efgh.slice"
// This function always expands the systemd name into the cgroupfs form. If only
// the last part is needed, use path.Base(...) on it to discard the rest.
func (cgroupName CgroupName) ToSystemd() string {
	if len(cgroupName) == 0 || (len(cgroupName) == 1 && cgroupName[0] == "") {
		return "/"
	}
	var newparts []string
	for _, part := range cgroupName {
		part = escapeSystemdCgroupName(part)
		newparts = append(newparts, part)
	}

	result, err := cgroupsystemd.ExpandSlice(strings.Join(newparts, "-") + systemdSuffix)
	if err != nil {
		// Should never happen...
		panic(fmt.Errorf("error converting cgroup name [%v] to systemd format: %v", cgroupName, err))
	}

	return result
}

func ParseSystemdToCgroupName(name string) CgroupName {
	driverName := path.Base(name)
	driverName = strings.TrimSuffix(driverName, systemdSuffix)
	parts := strings.Split(driverName, "-")
	var result []string
	for _, part := range parts {
		result = append(result, unescapeSystemdCgroupName(part))
	}
	return CgroupName(result)
}

func ParseCgroupfsToCgroupName(name string) CgroupName {
	components := strings.Split(strings.TrimPrefix(name, "/"), "/") // split("a/", "/")
	if len(components) == 1 && components[0] == "" {
		components = []string{}
	}

	klog.Infof("[ParseCgroupfsToCgroupName]components: %v", components)

	// "a", ""
	return CgroupName(components)
}

func NewCgroupName(base CgroupName, components ...string) CgroupName {
	for _, component := range components {
		if strings.Contains(component, "/") || strings.Contains(component, "_") {
			panic(fmt.Errorf("invalid character in component [%q] of CgroupName", component))
		}
	}

	return CgroupName(append(append([]string{}, base...), components...))
}

type libcontainerCgroupManagerType string

const (
	// libcontainerCgroupfs means use libcontainer with cgroupfs
	libcontainerCgroupfs libcontainerCgroupManagerType = "cgroupfs"
	// libcontainerSystemd means use libcontainer with systemd
	libcontainerSystemd libcontainerCgroupManagerType = "systemd"
	// systemdSuffix is the cgroup name suffix for systemd
	systemdSuffix string = ".slice"
)

// CgroupSubsystems holds information about the mounted cgroup subsystems
type CgroupSubsystems struct {
	// Cgroup subsystem mounts.
	// e.g.: "/sys/fs/cgroup/cpu" -> ["cpu", "cpuacct"]
	Mounts []libcontainercgroups.Mount

	// Cgroup subsystem to their mount location.
	// e.g.: "cpu" -> "/sys/fs/cgroup/cpu"
	MountPoints map[string]string
}

type libcontainerAdapter struct {
	// cgroupManagerType defines how to interface with libcontainer
	cgroupManagerType libcontainerCgroupManagerType
}

func (l *libcontainerAdapter) newManager(cgroups *libcontainerconfigs.Cgroup,
	paths map[string]string) (libcontainercgroups.Manager, error) {
	switch l.cgroupManagerType {
	case libcontainerCgroupfs:
		return cgroupfs.NewManager(cgroups, paths, false), nil
	}

	return nil, fmt.Errorf("invalid cgroup manager configuration")
}

func newLibcontainerAdapter(cgroupManagerType libcontainerCgroupManagerType) *libcontainerAdapter {
	return &libcontainerAdapter{cgroupManagerType: cgroupManagerType}
}

// It uses the Libcontainer raw fs cgroup manager for cgroup management.
type cgroupManagerImpl struct {
	// subsystems holds information about all the
	// mounted cgroup subsystems on the node
	subsystems *CgroupSubsystems
	// simplifies interaction with libcontainer and its cgroup managers
	adapter *libcontainerAdapter
}

func (cgroupManager *cgroupManagerImpl) toResources(resourceConfig *ResourceConfig) *libcontainerconfigs.Resources {
	resources := &libcontainerconfigs.Resources{
		Devices: []*libcontainerconfigs.DeviceRule{
			{
				Type:        'a',
				Permissions: "rwm",
				Allow:       true,
				Minor:       libcontainerconfigs.Wildcard,
				Major:       libcontainerconfigs.Wildcard,
			},
		},
		SkipDevices: true,
	}

	if resourceConfig == nil {
		return resources
	}
	if resourceConfig.Memory != nil {
		resources.Memory = *resourceConfig.Memory
	}
	if resourceConfig.CpuShares != nil {
		// cgroup v1
		resources.CpuShares = *resourceConfig.CpuShares
	}
	if resourceConfig.CpuQuota != nil {
		resources.CpuQuota = *resourceConfig.CpuQuota
	}
	if resourceConfig.CpuPeriod != nil {
		resources.CpuPeriod = *resourceConfig.CpuPeriod
	}
	if resourceConfig.PidsLimit != nil {
		resources.PidsLimit = *resourceConfig.PidsLimit
	}

	// if huge pages are enabled, we set them in libcontainer
	// for each page size enumerated, set that value
	pageSizes := sets.NewString()
	for pageSize, limit := range resourceConfig.HugePageLimit {
		sizeString, err := v1helper.HugePageUnitSizeFromByteSize(pageSize)
		if err != nil {
			klog.Warningf("pageSize is invalid: %v", err)
			continue
		}
		resources.HugetlbLimit = append(resources.HugetlbLimit, &libcontainerconfigs.HugepageLimit{
			Pagesize: sizeString,
			Limit:    uint64(limit),
		})
		pageSizes.Insert(sizeString)
	}
	// for each page size omitted, limit to 0
	for _, pageSize := range cgroupfs.HugePageSizes {
		if pageSizes.Has(pageSize) {
			continue
		}
		resources.HugetlbLimit = append(resources.HugetlbLimit, &libcontainerconfigs.HugepageLimit{
			Pagesize: pageSize,
			Limit:    uint64(0),
		})
	}

	return resources
}

// 创建一个cgroup
func (cgroupManager *cgroupManagerImpl) Create(cgroupConfig *CgroupConfig) error {
	resources := cgroupManager.toResources(cgroupConfig.ResourceParameters)
	libcontainerCgroupConfig := &libcontainerconfigs.Cgroup{
		Resources: resources,
	}
	// libcontainer consumes a different field and expects a different syntax
	// depending on the cgroup driver in use, so we need this conditional here.
	if cgroupManager.adapter.cgroupManagerType == libcontainerSystemd {
		updateSystemdCgroupInfo(libcontainerCgroupConfig, cgroupConfig.Name)
	} else {
		libcontainerCgroupConfig.Path = cgroupConfig.Name.ToCgroupfs()
	}

	libcontainerCgroupConfig.PidsLimit = *cgroupConfig.ResourceParameters.PidsLimit
	manager, err := cgroupManager.adapter.newManager(libcontainerCgroupConfig, nil)
	if err != nil {
		return err
	}

	// Apply(-1) is a hack to create the cgroup directories for each resource
	// subsystem. The function [cgroups.Manager.apply()] applies cgroup
	// configuration to the process with the specified pid.
	// It creates cgroup files for each subsystems and writes the pid
	// in the tasks file. We use the function to create all the required
	// cgroup files but not attach any "real" pid to the cgroup.
	if err := manager.Apply(-1); err != nil {
		return err
	}

	// it may confuse why we call set after we do apply, but the issue is that runc
	// follows a similar pattern.  it's needed to ensure cpu quota is set properly.
	if err := cgroupManager.Update(cgroupConfig); err != nil {
		utilruntime.HandleError(fmt.Errorf("cgroup update failed %v", err))
	}

	return nil
}

// TODO(filbranden): This logic belongs in libcontainer/cgroup/systemd instead.
// It should take a libcontainerconfigs.Cgroup.Path field (rather than Name and Parent)
// and split it appropriately, using essentially the logic below.
// This was done for cgroupfs in opencontainers/runc#497 but a counterpart
// for systemd was never introduced.
func updateSystemdCgroupInfo(cgroupConfig *libcontainerconfigs.Cgroup, cgroupName CgroupName) {
	dir, base := path.Split(cgroupName.ToSystemd())
	if dir == "/" {
		dir = "-.slice"
	} else {
		dir = path.Base(dir)
	}
	cgroupConfig.Parent = dir
	cgroupConfig.Name = base
}

func (cgroupManager *cgroupManagerImpl) Destroy(config *CgroupConfig) error {
	panic("implement me")
}

func (cgroupManager *cgroupManagerImpl) Update(config *CgroupConfig) error {
	panic("implement me")
}

func (cgroupManager *cgroupManagerImpl) Exists(name CgroupName) bool {
	panic("implement me")
}

func (cgroupManager *cgroupManagerImpl) Name(name CgroupName) string {
	panic("implement me")
}

func (cgroupManager *cgroupManagerImpl) CgroupName(name string) CgroupName {
	panic("implement me")
}

func (cgroupManager *cgroupManagerImpl) Pids(name CgroupName) []int {
	panic("implement me")
}

func (cgroupManager *cgroupManagerImpl) ReduceCPULimits(cgroupName CgroupName) error {
	panic("implement me")
}

func (cgroupManager *cgroupManagerImpl) GetResourceStats(name CgroupName) (*ResourceStats, error) {
	panic("implement me")
}

func NewCgroupManager(cgroupSubsystems *CgroupSubsystems, cgroupDriver string) CgroupManager {
	managerType := libcontainerCgroupfs
	if cgroupDriver == string(libcontainerSystemd) {
		managerType = libcontainerSystemd
	}
	return &cgroupManagerImpl{
		subsystems: cgroupSubsystems,
		adapter:    newLibcontainerAdapter(managerType),
	}
}
