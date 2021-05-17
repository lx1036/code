package dockershim

import (
	"fmt"
	"strings"

	"k8s-lx1036/k8s/kubelet/pkg/types"

	dockertypes "github.com/docker/docker/api/types"
	dockerfilters "github.com/docker/docker/api/types/filters"
)

const (
	annotationPrefix     = "annotation."
	securityOptSeparator = '='

	dockerNetNSFmt = "/proc/%v/ns/net"
)

var internalLabelKeys = []string{containerTypeLabelKey, containerLogPathLabelKey, sandboxIDLabelKey}

// dockerFilter wraps around dockerfilters.Args and provides methods to modify
// the filter easily.
type dockerFilter struct {
	args *dockerfilters.Args
}

func newDockerFilter(args *dockerfilters.Args) *dockerFilter {
	return &dockerFilter{args: args}
}

func (f *dockerFilter) Add(key, value string) {
	f.args.Add(key, value)
}

func (f *dockerFilter) AddLabel(key, value string) {
	f.Add("label", fmt.Sprintf("%s=%s", key, value))
}

/*
"Labels":{
	"annotation.io.kubernetes.container.hash":"ad01c60e",
	"annotation.io.kubernetes.container.restartCount":"0",
	"annotation.io.kubernetes.container.terminationMessagePath":"/dev/termination-log",
	"annotation.io.kubernetes.container.terminationMessagePolicy":"File",
	"annotation.io.kubernetes.pod.terminationGracePeriod":"30",
	"io.kubernetes.container.logpath":"/var/log/pods/default_cgroup1-75cb7bc8c5-vbzww_cf1f7aa0-acb5-48cf-a2e6-50d34d553d96/cgroup1-0/0.log",
	"io.kubernetes.container.name":"cgroup1-0",
	"io.kubernetes.docker.type":"container",
	"io.kubernetes.pod.name":"cgroup1-75cb7bc8c5-vbzww",
	"io.kubernetes.pod.namespace":"default",
	"io.kubernetes.pod.uid":"cf1f7aa0-acb5-48cf-a2e6-50d34d553d96",
	"io.kubernetes.sandbox.id":"af88d9ab332141c168be34dc49752787d0640f0877b181d94077e24fcf9de497"
}
*/
// extractLabels converts raw docker labels to the CRI labels and annotations.
// It also filters out internal labels used by this shim.
func extractLabels(input map[string]string) (map[string]string, map[string]string) {
	labels := make(map[string]string)
	annotations := make(map[string]string)
	for k, v := range input {
		// Check if the key is used internally by the shim.
		internal := false
		for _, internalKey := range internalLabelKeys {
			if k == internalKey {
				internal = true
				break
			}
		}
		if internal {
			continue
		}

		// Delete the container name label for the sandbox. It is added in the shim,
		// should not be exposed via CRI.
		if k == types.KubernetesContainerNameLabel &&
			input[containerTypeLabelKey] == containerTypeLabelSandbox {
			continue
		}

		// Check if the label should be treated as an annotation.
		if strings.HasPrefix(k, annotationPrefix) {
			annotations[strings.TrimPrefix(k, annotationPrefix)] = v
			continue
		}
		labels[k] = v
	}
	return labels, annotations
}

func getNetworkNamespace(c *dockertypes.ContainerJSON) (string, error) {
	if c.State.Pid == 0 {
		// Docker reports pid 0 for an exited container.
		return "", fmt.Errorf("cannot find network namespace for the terminated container %q", c.ID)
	}

	return fmt.Sprintf(dockerNetNSFmt, c.State.Pid), nil
}
