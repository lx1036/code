package cgroup

import (
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/defaults"
)

var (
	// Path to where cgroup is mounted
	cgroupRoot = defaults.DefaultCgroupRoot
)

// GetCgroupRoot returns the path for the cgroupv2 mount
func GetCgroupRoot() string {
	return cgroupRoot
}
