package server

import (
	"k8s.io/kubernetes/pkg/kubelet/cm"
)

// INFO: @see https://github.com/kubernetes/kubernetes/blob/release-1.19/cmd/kubelet/app/server.go#L604-L606
func (server *Server) GetCgroupRoots() []string {
	// s.CgroupRoot="/" CgroupsPerQOS=true CgroupDriver="cgroupfs"
	var cgroupRoots []string
	nodeAllocatableRoot := cm.NodeAllocatableRoot(server.CgroupRoot, server.CgroupsPerQOS, server.CgroupDriver)
	cgroupRoots = append(cgroupRoots, nodeAllocatableRoot) // "/kubepods"

	return cgroupRoots
}
