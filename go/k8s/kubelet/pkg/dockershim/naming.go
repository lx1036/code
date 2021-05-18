package dockershim

import (
	"fmt"
	"strconv"
	"strings"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

const (
	// kubePrefix is used to identify the containers/sandboxes on the node managed by kubelet
	kubePrefix = "k8s"

	// Delimiter used to construct docker container names.
	nameDelimiter = "_"
	// DockerImageIDPrefix is the prefix of image id in container status.
	DockerImageIDPrefix = "docker://"
	// DockerPullableImageIDPrefix is the prefix of pullable image id in container status.
	DockerPullableImageIDPrefix = "docker-pullable://"
)

func makeContainerName(s *runtimeapi.PodSandboxConfig, c *runtimeapi.ContainerConfig) string {
	return strings.Join([]string{
		kubePrefix,                            // 0
		c.Metadata.Name,                       // 1:
		s.Metadata.Name,                       // 2: sandbox name
		s.Metadata.Namespace,                  // 3: sandbox namesapce
		s.Metadata.Uid,                        // 4  sandbox uid
		fmt.Sprintf("%d", c.Metadata.Attempt), // 5
	}, nameDelimiter)
}

func makeSandboxName(s *runtimeapi.PodSandboxConfig) string {
	return strings.Join([]string{
		kubePrefix,                            // 0
		PodInfraContainerName,                 // 1
		s.Metadata.Name,                       // 2
		s.Metadata.Namespace,                  // 3
		s.Metadata.Uid,                        // 4
		fmt.Sprintf("%d", s.Metadata.Attempt), // 5
	}, nameDelimiter)
}

// INFO: name="/k8s_cgroup1-0_cgroup1-75cb7bc8c5-vbzww_default_cf1f7aa0-acb5-48cf-a2e6-50d34d553d96_0"
func parseSandboxName(name string) (*runtimeapi.PodSandboxMetadata, error) {
	// Docker adds a "/" prefix to names. so trim it.
	name = strings.TrimPrefix(name, "/")

	parts := strings.Split(name, nameDelimiter)
	// Tolerate the random suffix.
	// TODO(random-liu): Remove 7 field case when docker 1.11 is deprecated.
	if len(parts) != 6 && len(parts) != 7 {
		return nil, fmt.Errorf("failed to parse the sandbox name: %q", name)
	}
	if parts[0] != kubePrefix {
		return nil, fmt.Errorf("container is not managed by kubernetes: %q", name)
	}

	attempt, err := parseUint32(parts[5])
	if err != nil {
		return nil, fmt.Errorf("failed to parse the sandbox name %q: %v", name, err)
	}

	return &runtimeapi.PodSandboxMetadata{
		Name:      parts[2],
		Namespace: parts[3],
		Uid:       parts[4],
		Attempt:   attempt,
	}, nil
}

func parseUint32(s string) (uint32, error) {
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}
