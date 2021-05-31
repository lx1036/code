package nodeports

import (
	"context"
	"fmt"

	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = "NodePorts"

	// preFilterStateKey is the key in CycleState to NodePorts pre-computed data.
	// Using the name of the plugin will likely help us avoid collisions with other plugins.
	preFilterStateKey = "PreFilter" + Name

	ErrReason = "node(s) didn't have free ports for the requested pod ports"
)

type preFilterState []*v1.ContainerPort

func (s preFilterState) Clone() framework.StateData {
	// The state is not impacted by adding/removing existing pods, hence we don't need to make a deep copy.
	return s
}

// NodePorts is a plugin that checks if a node has free ports for the requested pod ports.
type NodePorts struct{}

func (pl *NodePorts) Name() string {
	return Name
}

func (pl *NodePorts) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (pl *NodePorts) PreFilter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod) *framework.Status {
	s := getContainerPorts(pod)
	cycleState.Write(preFilterStateKey, preFilterState(s)) // 保存数据给 Filter plugin 使用

	return nil
}

func (pl *NodePorts) Filter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod,
	nodeInfo *framework.NodeInfo) *framework.Status {
	wantPorts, err := getPreFilterState(cycleState)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}

	fits := fitsPorts(wantPorts, nodeInfo)
	if !fits {
		return framework.NewStatus(framework.Unschedulable, ErrReason)
	}

	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, _ framework.FrameworkHandle) (framework.Plugin, error) {
	return &NodePorts{}, nil
}

func getContainerPorts(pods ...*v1.Pod) []*v1.ContainerPort {
	ports := []*v1.ContainerPort{}
	for _, pod := range pods {
		for j := range pod.Spec.Containers {
			container := &pod.Spec.Containers[j]
			for k := range container.Ports {
				ports = append(ports, &container.Ports[k])
			}
		}
	}
	return ports
}

func getPreFilterState(cycleState *framework.CycleState) (preFilterState, error) {
	c, err := cycleState.Read(preFilterStateKey)
	if err != nil {
		// preFilterState doesn't exist, likely PreFilter wasn't invoked.
		return nil, fmt.Errorf("error reading %q from cycleState: %v", preFilterStateKey, err)
	}

	s, ok := c.(preFilterState)
	if !ok {
		return nil, fmt.Errorf("%+v  convert to nodeports.preFilterState error", c)
	}
	return s, nil
}

func fitsPorts(wantPorts []*v1.ContainerPort, nodeInfo *framework.NodeInfo) bool {
	// try to see whether existingPorts and wantPorts will conflict or not
	existingPorts := nodeInfo.UsedPorts
	for _, cp := range wantPorts {
		if existingPorts.CheckConflict(cp.HostIP, string(cp.Protocol), cp.HostPort) {
			return false
		}
	}
	return true
}
