package dockershim

import (
	"fmt"
	"k8s.io/kubernetes/pkg/kubelet/dockershim/libdocker"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

func containerToRuntimeAPISandbox(c *dockertypes.Container) (*runtimeapi.PodSandbox, error) {
	state := toRuntimeAPISandboxState(c.Status)
	if len(c.Names) == 0 {
		return nil, fmt.Errorf("unexpected empty sandbox name: %+v", c)
	}
	metadata, err := parseSandboxName(c.Names[0])
	if err != nil {
		return nil, err
	}
	labels, annotations := extractLabels(c.Labels)
	// The timestamp in dockertypes.Container is in seconds.
	createdAt := c.Created * int64(time.Second)

	return &runtimeapi.PodSandbox{
		Id:          c.ID,
		Metadata:    metadata,
		State:       state,
		CreatedAt:   createdAt,
		Labels:      labels,
		Annotations: annotations,
	}, nil
}

func toRuntimeAPISandboxState(state string) runtimeapi.PodSandboxState {
	// Parse the state string in dockertypes.Container. This could break when
	// we upgrade docker.
	switch {
	case strings.HasPrefix(state, libdocker.StatusRunningPrefix):
		return runtimeapi.PodSandboxState_SANDBOX_READY
	default:
		return runtimeapi.PodSandboxState_SANDBOX_NOTREADY
	}
}

func checkpointToRuntimeAPISandbox(id string, checkpoint DockershimCheckpoint) *runtimeapi.PodSandbox {
	state := runtimeapi.PodSandboxState_SANDBOX_NOTREADY
	_, name, namespace, _, _ := checkpoint.GetData()
	return &runtimeapi.PodSandbox{
		Id: id,
		Metadata: &runtimeapi.PodSandboxMetadata{
			Name:      name,
			Namespace: namespace,
		},
		State: state,
	}
}
