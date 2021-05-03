package testing

import (
	kubecontainer "k8s-lx1036/k8s/kubelet/pkg/container"

	"github.com/stretchr/testify/mock"

	v1 "k8s.io/api/core/v1"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) PullImage(image kubecontainer.ImageSpec, pullSecrets []v1.Secret, podSandboxConfig *runtimeapi.PodSandboxConfig) (string, error) {
	panic("implement me")
}

func (m *Mock) GetImageRef(image kubecontainer.ImageSpec) (string, error) {
	panic("implement me")
}

func (m *Mock) ListImages() ([]kubecontainer.Image, error) {
	panic("implement me")
}

func (m *Mock) RemoveImage(image kubecontainer.ImageSpec) error {
	panic("implement me")
}

func (m *Mock) ImageStats() (*kubecontainer.ImageStats, error) {
	panic("implement me")
}
