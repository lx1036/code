package testing

import (
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/events"
	cadvisorapi "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	cadvisorapiv2 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"

	"github.com/stretchr/testify/mock"
)

// Mock cadvisor.Interface implementation.
type Mock struct {
	mock.Mock
}

func (c *Mock) Start() error {
	args := c.Called()
	return args.Error(0)
}

func (c *Mock) DockerContainer(name string, req *cadvisorapi.ContainerInfoRequest) (cadvisorapi.ContainerInfo, error) {
	args := c.Called(name, req)
	return args.Get(0).(cadvisorapi.ContainerInfo), args.Error(1)
}

func (c *Mock) ContainerInfo(name string, req *cadvisorapi.ContainerInfoRequest) (*cadvisorapi.ContainerInfo, error) {
	panic("implement me")
}

func (c *Mock) ContainerInfoV2(name string, options cadvisorapiv2.RequestOptions) (map[string]cadvisorapiv2.ContainerInfo, error) {
	panic("implement me")
}

func (c *Mock) GetRequestedContainersInfo(containerName string, options cadvisorapiv2.RequestOptions) (map[string]*cadvisorapi.ContainerInfo, error) {
	panic("implement me")
}

func (c *Mock) SubcontainerInfo(name string, req *cadvisorapi.ContainerInfoRequest) (map[string]*cadvisorapi.ContainerInfo, error) {
	panic("implement me")
}

func (c *Mock) MachineInfo() (*cadvisorapi.MachineInfo, error) {
	panic("implement me")
}

func (c *Mock) VersionInfo() (*cadvisorapi.VersionInfo, error) {
	panic("implement me")
}

func (c *Mock) ImagesFsInfo() (cadvisorapiv2.FsInfo, error) {
	panic("implement me")
}

func (c *Mock) RootFsInfo() (cadvisorapiv2.FsInfo, error) {
	panic("implement me")
}

func (c *Mock) WatchEvents(request *events.Request) (*events.EventChannel, error) {
	panic("implement me")
}

func (c *Mock) GetDirFsInfo(path string) (cadvisorapiv2.FsInfo, error) {
	panic("implement me")
}
