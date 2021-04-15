package devicemanager

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// endpoint maps to a single registered device plugin. It is responsible
// for managing gRPC communications with the device plugin and caching
// device states reported by the device plugin.
type Endpoint interface {
	run()
	stop()
	getPreferredAllocation(available, mustInclude []string, size int) (*pluginapi.PreferredAllocationResponse, error)
	allocate(devs []string) (*pluginapi.AllocateResponse, error)
	preStartContainer(devs []string) (*pluginapi.PreStartContainerResponse, error)
	callback(resourceName string, devices []pluginapi.Device)
	isStopped() bool
	stopGracePeriodExpired() bool
}

type endpointImpl struct {
	client     pluginapi.DevicePluginClient
	clientConn *grpc.ClientConn

	socketPath   string
	resourceName string
	stopTime     time.Time

	mutex sync.Mutex
	cb    monitorCallback
}

// INFO: grpc client，用来 plugin registration
func newEndpointImpl(socketPath, resourceName string, callback monitorCallback) (*endpointImpl, error) {
	client, c, err := dial(socketPath)
	if err != nil {
		klog.Errorf("Can't create new endpoint with path %s err %v", socketPath, err)
		return nil, err
	}

	return &endpointImpl{
		client:     client,
		clientConn: c,

		socketPath:   socketPath,
		resourceName: resourceName,

		cb: callback,
	}, nil
}

// newStoppedEndpointImpl creates a new endpoint for the given resourceName with stopTime set.
// This is to be used during Kubelet restart, before the actual device plugin re-registers.
func newStoppedEndpointImpl(resourceName string) *endpointImpl {
	return &endpointImpl{
		resourceName: resourceName,
		stopTime:     time.Now(),
	}
}

// run initializes ListAndWatch gRPC call for the device plugin and
// blocks on receiving ListAndWatch gRPC stream updates. Each ListAndWatch
// stream update contains a new list of device states.
// It then issues a callback to pass this information to the device manager which
// will adjust the resource available information accordingly.
func (endpoint *endpointImpl) run() {
	stream, err := endpoint.client.ListAndWatch(context.Background(), &pluginapi.Empty{})
	if err != nil {
		klog.Errorf("listAndWatch ended unexpectedly for device plugin %s with error %v", endpoint.resourceName, err)

		return
	}

	for {
		response, err := stream.Recv()
		if err != nil {
			klog.Errorf("listAndWatch ended unexpectedly for device plugin %s with error %v", endpoint.resourceName, err)
			return
		}

		devs := response.Devices
		klog.V(2).Infof("State pushed for device plugin %s", endpoint.resourceName)

		var newDevs []pluginapi.Device
		for _, d := range devs {
			newDevs = append(newDevs, *d)
		}

		endpoint.callback(endpoint.resourceName, newDevs)
	}
}
func (endpoint *endpointImpl) stop() {}
func (endpoint *endpointImpl) getPreferredAllocation(available, mustInclude []string, size int) (*pluginapi.PreferredAllocationResponse, error) {
	return nil, nil
}

// allocate issues Allocate gRPC call to the device plugin.
func (endpoint *endpointImpl) allocate(devs []string) (*pluginapi.AllocateResponse, error) {
	if endpoint.isStopped() {
		return nil, fmt.Errorf("endpoint %v has been stopped", endpoint)
	}

	return endpoint.client.Allocate(context.Background(), &pluginapi.AllocateRequest{
		ContainerRequests: []*pluginapi.ContainerAllocateRequest{
			{DevicesIDs: devs},
		},
	})
}
func (endpoint *endpointImpl) preStartContainer(devs []string) (*pluginapi.PreStartContainerResponse, error) {
	return nil, nil
}
func (endpoint *endpointImpl) callback(resourceName string, devices []pluginapi.Device) {
	endpoint.cb(resourceName, devices)
}
func (endpoint *endpointImpl) isStopped() bool {
	endpoint.mutex.Lock()
	defer endpoint.mutex.Unlock()

	return !endpoint.stopTime.IsZero()
}
func (endpoint *endpointImpl) stopGracePeriodExpired() bool {
	return false
}

// dial establishes the gRPC communication with the registered device plugin. https://godoc.org/google.golang.org/grpc#Dial
func dial(unixSocketPath string) (pluginapi.DevicePluginClient, *grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := grpc.DialContext(ctx, unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial device plugin: %v", err)
	}

	return pluginapi.NewDevicePluginClient(c), c, nil
}
