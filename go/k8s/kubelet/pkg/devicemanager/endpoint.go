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
type endpoint interface {
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

// newEndpointImpl creates a new endpoint for the given resourceName.
// This is to be used during normal device plugin registration.
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

func (e *endpointImpl) run() {
}
func (e *endpointImpl) stop() {}
func (e *endpointImpl) getPreferredAllocation(available, mustInclude []string, size int) (*pluginapi.PreferredAllocationResponse, error) {
	return nil, nil
}
func (e *endpointImpl) allocate(devs []string) (*pluginapi.AllocateResponse, error) {
	return nil, nil
}
func (e *endpointImpl) preStartContainer(devs []string) (*pluginapi.PreStartContainerResponse, error) {
	return nil, nil
}
func (e *endpointImpl) callback(resourceName string, devices []pluginapi.Device) {}
func (e *endpointImpl) isStopped() bool {
	return false
}
func (e *endpointImpl) stopGracePeriodExpired() bool {
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
