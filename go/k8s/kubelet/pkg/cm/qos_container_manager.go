package cm

import v1 "k8s.io/api/core/v1"

type ActivePodsFunc func() []*v1.Pod

type QOSContainerManager interface {
	Start(func() v1.ResourceList, ActivePodsFunc) error
	GetQOSContainersInfo() QOSContainersInfo
	UpdateCgroups() error
}
