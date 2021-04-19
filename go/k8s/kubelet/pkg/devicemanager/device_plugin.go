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

// INFO: DevicePlugin 对象可以作为一个注册插件到kubelet中的注册框架，
// 比如写一个自己的 gpu device plugin

// INFO: 插件通过 Unix socket 在主机路径 /var/lib/kubelet/device-plugins/kubelet.sock 处向 kubelet 注册自身

// Stub implementation for DevicePlugin.
type DevicePlugin struct {
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

// INFO: grpc server
func NewDevicePlugin(devs []*pluginapi.Device, socket string, name string, preStartContainerFlag bool, getPreferredAllocationFlag bool) *DevicePlugin {
	return &DevicePlugin{
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

func (devicePlugin *DevicePlugin) cleanup() error {
	if err := os.Remove(devicePlugin.socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (devicePlugin *DevicePlugin) GetInfo(context context.Context,
	request *watcherapi.InfoRequest) (*watcherapi.PluginInfo, error) {
	panic("implement me")
}

func (devicePlugin *DevicePlugin) NotifyRegistrationStatus(context context.Context,
	status *watcherapi.RegistrationStatus) (*watcherapi.RegistrationStatusResponse, error) {
	panic("implement me")
}

func (devicePlugin *DevicePlugin) GetDevicePluginOptions(ctx context.Context,
	empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	panic("implement me")
}

// ListAndWatch 返回 Device 列表构成的数据流。
// 当 Device 状态发生变化或者 Device 消失时，ListAndWatch 会返回新的列表。
func (devicePlugin *DevicePlugin) ListAndWatch(empty *pluginapi.Empty,
	server pluginapi.DevicePlugin_ListAndWatchServer) error {
	klog.Info("ListAndWatch")

	server.Send(&pluginapi.ListAndWatchResponse{Devices: devicePlugin.devs})

	for {
		select {
		case <-devicePlugin.stop:
			return nil
		case updated := <-devicePlugin.update:
			server.Send(&pluginapi.ListAndWatchResponse{Devices: updated})
		}
	}
}

// GetPreferredAllocation 从一组可用的设备中返回一些优选的设备用来分配，
// 所返回的优选分配结果不一定会是设备管理器的最终分配方案。
// 此接口的设计仅是为了让设备管理器能够在可能的情况下做出更有意义的决定。
func (devicePlugin *DevicePlugin) GetPreferredAllocation(ctx context.Context,
	request *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return nil, nil
}

// INFO: grpc server 的 Allocate()，很重要!!!
// Allocate 在容器创建期间调用，这样设备插件可以运行一些特定于设备的操作，
// 并告诉 kubelet 如何令 Device 可在容器中访问所需执行的具体步骤
func (devicePlugin *DevicePlugin) Allocate(ctx context.Context,
	request *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	klog.Infof("[Allocate] %+v", request)

	devs := make(map[string]pluginapi.Device)

	for _, dev := range devicePlugin.devs {
		devs[dev.ID] = *dev
	}

	return devicePlugin.allocFunc(request, devs)
}

// SetAllocFunc sets allocFunc of the device plugin
func (devicePlugin *DevicePlugin) SetAllocFunc(f stubAllocFunc) {
	devicePlugin.allocFunc = f
}

// PreStartContainer 在设备插件注册阶段根据需要被调用，调用发生在容器启动之前。
// 在将设备提供给容器使用之前，设备插件可以运行一些诸如重置设备之类的特定于
// 具体设备的操作
func (devicePlugin *DevicePlugin) PreStartContainer(ctx context.Context,
	request *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return nil, nil
}

// Update allows the device plugin to send new devices through ListAndWatch
func (devicePlugin *DevicePlugin) Update(devs []*pluginapi.Device) {
	devicePlugin.update <- devs
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (devicePlugin *DevicePlugin) Register(kubeletEndpoint, resourceName string, pluginSockDir string) error {

	return nil
}

// Start starts the gRPC server of the device plugin. Can only
// be called once.
func (devicePlugin *DevicePlugin) Start() error {
	err := devicePlugin.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", devicePlugin.socket)
	if err != nil {
		return err
	}

	devicePlugin.wg.Add(1)
	devicePlugin.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(devicePlugin.server, devicePlugin)
	watcherapi.RegisterRegistrationServer(devicePlugin.server, devicePlugin)

	go func() {
		defer devicePlugin.wg.Done()
		devicePlugin.server.Serve(sock)
	}()

	// 测试 grpc /tmp/device-plugin.sock 是否连通
	_, conn, err := dial(devicePlugin.socket)
	if err != nil {
		return err
	}
	conn.Close()
	klog.Infof("Starting to serve on %v", devicePlugin.socket)

	return nil
}

// 关闭 grpc server，并清理 socket 文件，可以多次调用Stop()
func (devicePlugin *DevicePlugin) Stop() error {
	if devicePlugin.server == nil {
		return nil
	}
	devicePlugin.server.Stop() // devicePlugin.server.Serve(sock)开始阻塞，Stop()后会跳出阻塞
	devicePlugin.wg.Wait()     // goroutine 内会 defer devicePlugin.wg.Done()
	devicePlugin.server = nil
	close(devicePlugin.stop) // This prevents re-starting the server.

	return devicePlugin.cleanup()
}
