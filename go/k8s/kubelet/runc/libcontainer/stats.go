package libcontainer

import (
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups"
	"k8s-lx1036/k8s/kubelet/runc/types"
)

type Stats struct {
	Interfaces  []*types.NetworkInterface
	CgroupStats *cgroups.Stats
}
