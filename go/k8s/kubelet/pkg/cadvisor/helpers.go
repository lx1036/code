package cadvisor

import (
	"fmt"
	cadvisorfs "github.com/google/cadvisor/fs"
	"k8s.io/kubernetes/pkg/kubelet/types"
)

// imageFsInfoProvider knows how to translate the configured runtime
// to its file system label for images.
type imageFsInfoProvider struct {
	runtime         string
	runtimeEndpoint string
}

func (i imageFsInfoProvider) ImageFsInfoLabel() (string, error) {
	switch i.runtime {
	case types.DockerContainerRuntime:
		return cadvisorfs.LabelDockerImages, nil
	}
	
	return "", fmt.Errorf("no imagefs label for configured runtime")
}

// NewImageFsInfoProvider returns a provider for the specified runtime configuration.
func NewImageFsInfoProvider(runtime, runtimeEndpoint string) ImageFsInfoProvider {
	return &imageFsInfoProvider{runtime: runtime, runtimeEndpoint: runtimeEndpoint}
}
