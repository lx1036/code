package state

import "k8s-lx1036/k8s/kubelet/pkg/cm/containermap"

func NewCheckpointState(stateDir, checkpointName, policyName string,
	initialContainers containermap.ContainerMap) (State, error) {

	return nil, nil
}
