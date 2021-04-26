package testing

import (
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"

	"github.com/stretchr/testify/mock"
)

// This struct mocks a container handler.
type MockContainerHandler struct {
	mock.Mock
	Name    string
	Aliases []string
}

func NewMockContainerHandler(containerName string) *MockContainerHandler {
	return &MockContainerHandler{
		Name: containerName,
	}
}

// If self.Name is not empty, then ContainerReference() will return self.Name and self.Aliases.
// Otherwise, it will use the value provided by .On().Return().
func (h *MockContainerHandler) ContainerReference() (v1.ContainerReference, error) {
	if len(h.Name) > 0 {
		var aliases []string
		if len(h.Aliases) > 0 {
			aliases = make([]string, len(h.Aliases))
			copy(aliases, h.Aliases)
		}
		return v1.ContainerReference{
			Name:    h.Name,
			Aliases: aliases,
		}, nil
	}
	args := h.Called()
	return args.Get(0).(v1.ContainerReference), args.Error(1)
}

func (h *MockContainerHandler) Start() {}

func (h *MockContainerHandler) Cleanup() {}

func (h *MockContainerHandler) GetSpec() (v1.ContainerSpec, error) {
	args := h.Called()
	return args.Get(0).(v1.ContainerSpec), args.Error(1)
}

func (h *MockContainerHandler) GetStats() (*v1.ContainerStats, error) {
	args := h.Called()
	return args.Get(0).(*v1.ContainerStats), args.Error(1)
}

func (h *MockContainerHandler) ListContainers(listType container.ListType) ([]v1.ContainerReference, error) {
	args := h.Called(listType)
	return args.Get(0).([]v1.ContainerReference), args.Error(1)
}

func (h *MockContainerHandler) ListProcesses(listType container.ListType) ([]int, error) {
	args := h.Called(listType)
	return args.Get(0).([]int), args.Error(1)
}

func (h *MockContainerHandler) Exists() bool {
	args := h.Called()
	return args.Get(0).(bool)
}

func (h *MockContainerHandler) GetCgroupPath(path string) (string, error) {
	args := h.Called(path)
	return args.Get(0).(string), args.Error(1)
}

func (h *MockContainerHandler) GetContainerLabels() map[string]string {
	args := h.Called()
	return args.Get(0).(map[string]string)
}

func (h *MockContainerHandler) Type() container.ContainerType {
	args := h.Called()
	return args.Get(0).(container.ContainerType)
}

func (h *MockContainerHandler) GetContainerIPAddress() string {
	args := h.Called()
	return args.Get(0).(string)
}
