package cm

import (
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

// INFO: pod level cgroups

const (
	podCgroupNamePrefix = "pod"
)

// QOSContainersInfo stores the names of containers per qos
type QOSContainersInfo struct {
	Guaranteed CgroupName
	BestEffort CgroupName
	Burstable  CgroupName
}

type podContainerManagerImpl struct {
	// qosContainersInfo hold absolute paths of the top level qos containers
	qosContainersInfo QOSContainersInfo
	// Stores the mounted cgroup subsystems
	subsystems *CgroupSubsystems
	// cgroupManager is the cgroup Manager Object responsible for managing all
	// pod cgroups.
	cgroupManager *cgroupManagerImpl
	// Maximum number of pids in a pod
	podPidsLimit int64
	// enforceCPULimits controls whether cfs quota is enforced or not
	enforceCPULimits bool
	// cpuCFSQuotaPeriod is the cfs period value, cfs_period_us, setting per
	// node for all containers in usec
	cpuCFSQuotaPeriod uint64
}

// IsPodCgroup returns true if the literal cgroupfs name corresponds to a pod
func (m *podContainerManagerImpl) IsPodCgroup(cgroupfs string) (bool, types.UID) {
	cgroupName := m.cgroupManager.CgroupName(cgroupfs)
	qosContainersList := [3]CgroupName{
		m.qosContainersInfo.BestEffort,
		m.qosContainersInfo.Burstable,
		m.qosContainersInfo.Guaranteed,
	}
	basePath := ""
	for _, qosContainerName := range qosContainersList {
		// a pod cgroup is a direct child of a qos node, so check if its a match
		if len(cgroupName) == len(qosContainerName)+1 {
			basePath = cgroupName[len(qosContainerName)]
		}
	}

	if basePath == "" {
		return false, types.UID("")
	}
	if !strings.HasPrefix(basePath, podCgroupNamePrefix) {
		return false, types.UID("")
	}

	parts := strings.Split(basePath, podCgroupNamePrefix)
	if len(parts) != 2 {
		return false, types.UID("")
	}

	// pod8dbc9e11-dc93-4eb2-bbef-ed21cf1db420 -> pod, 8dbc9e11-dc93-4eb2-bbef-ed21cf1db420
	return true, types.UID(parts[1])
}
