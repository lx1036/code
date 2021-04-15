package devicemanager

import (
	"context"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	watcherapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

// Stub implementation for DevicePlugin.
type Stub struct {
	devs                       []*pluginapi.Device
	socket                     string
	resourceName               string
	preStartContainerFlag      bool
	getPreferredAllocationFlag bool

	stop   chan interface{}
	wg     sync.WaitGroup
	update chan []*pluginapi.Device

	server *grpc.Server

	// allocFunc is used for handling allocation request
	allocFunc stubAllocFunc

	// getPreferredAllocFunc is used for handling getPreferredAllocation request
	getPreferredAllocFunc stubGetPreferredAllocFunc

	registrationStatus chan watcherapi.RegistrationStatus // for testing
	endpoint           string                             // for testing

}

// stubAllocFunc is the function called when an allocation request is received from Kubelet
type stubAllocFunc func(r *pluginapi.AllocateRequest, devs map[string]pluginapi.Device) (*pluginapi.AllocateResponse, error)

func defaultAllocFunc(r *pluginapi.AllocateRequest, devs map[string]pluginapi.Device) (*pluginapi.AllocateResponse, error) {
	var response pluginapi.AllocateResponse

	return &response, nil
}

// stubGetPreferredAllocFunc is the function called when a getPreferredAllocation request is received from Kubelet
type stubGetPreferredAllocFunc func(r *pluginapi.PreferredAllocationRequest,
	devs map[string]pluginapi.Device) (*pluginapi.PreferredAllocationResponse, error)

func defaultGetPreferredAllocFunc(r *pluginapi.PreferredAllocationRequest, devs map[string]pluginapi.Device) (*pluginapi.PreferredAllocationResponse, error) {
	var response pluginapi.PreferredAllocationResponse

	return &response, nil
}

// NewDevicePluginStub returns an initialized DevicePlugin Stub.
func NewDevicePluginStub(devs []*pluginapi.Device, socket string, name string, preStartContainerFlag bool, getPreferredAllocationFlag bool) *Stub {
	return &Stub{
		devs:                       devs,
		socket:                     socket,
		resourceName:               name,
		preStartContainerFlag:      preStartContainerFlag,
		getPreferredAllocationFlag: getPreferredAllocationFlag,

		stop:   make(chan interface{}),
		update: make(chan []*pluginapi.Device),

		allocFunc:             defaultAllocFunc,
		getPreferredAllocFunc: defaultGetPreferredAllocFunc,
	}
}

func (m *Stub) cleanup() error {
	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Start starts the gRPC server of the device plugin. Can only
// be called once.
func (m *Stub) Start() error {
	err := m.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}

	m.wg.Add(1)
	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(m.server, m)
	watcherapi.RegisterRegistrationServer(m.server, m)

	go func() {
		defer m.wg.Done()
		m.server.Serve(sock)
	}()

	// 测试 grpc /tmp/device-plugin.sock 是否连通
	_, conn, err := dial(m.socket)
	if err != nil {
		return err
	}
	conn.Close()
	klog.Infof("Starting to serve on %v", m.socket)

	return nil
}

func (m *Stub) GetInfo(context context.Context, request *watcherapi.InfoRequest) (*watcherapi.PluginInfo, error) {
	panic("implement me")
}

func (m *Stub) NotifyRegistrationStatus(context context.Context, status *watcherapi.RegistrationStatus) (*watcherapi.RegistrationStatusResponse, error) {
	panic("implement me")
}

func (m *Stub) GetDevicePluginOptions(ctx context.Context, empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	panic("implement me")
}

func (m *Stub) ListAndWatch(empty *pluginapi.Empty, server pluginapi.DevicePlugin_ListAndWatchServer) error {
	panic("implement me")
}

func (m *Stub) GetPreferredAllocation(ctx context.Context, request *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	panic("implement me")
}

func (m *Stub) Allocate(ctx context.Context, request *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	panic("implement me")
}

func (m *Stub) PreStartContainer(ctx context.Context, request *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	panic("implement me")
}

// 关闭 grpc server，并清理 socket 文件，可以多次调用Stop()
func (m *Stub) Stop() error {
	if m.server == nil {
		return nil
	}
	m.server.Stop() // m.server.Serve(sock)开始阻塞，Stop()后会跳出阻塞
	m.wg.Wait()     // goroutine 内会 defer m.wg.Done()
	m.server = nil
	close(m.stop) // This prevents re-starting the server.

	return m.cleanup()
}
