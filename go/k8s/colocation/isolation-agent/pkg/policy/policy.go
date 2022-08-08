package policy

import (
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

func GetCPUSetOrDefault(ratio float64) cpuset.CPUSet {

	return cpuset.NewCPUSet(0, 13)
}
