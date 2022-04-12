package deviceplugin

import (
	"context"
	"fmt"
	"net"
	"path"
	"regexp"
	"sync"
	"time"

	"google.golang.org/grpc"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// define resource name
const (
	ENITypeENI    = "eni"
	ENITypeMember = "member"

	// ENIResName vm eni resource name in kubernetes container resource
	ENIResName       = "vm/eni"
	MemberENIResName = "vm/member-eni"
)

type eniRes struct {
	resName string
	re      *regexp.Regexp
	sock    string
}

var eniMap = map[string]eniRes{
	ENITypeENI: {
		resName: ENIResName,
		re:      regexp.MustCompile("^.*" + "-eni.sock"),
		sock:    pluginapi.DevicePluginPath + "%d-" + "eni.sock",
	},
	ENITypeMember: {
		resName: MemberENIResName,
		re:      regexp.MustCompile("^.*" + "-member-eni.sock"),
		sock:    pluginapi.DevicePluginPath + "%d-" + "member-eni.sock",
	},
}

// ENIDevicePlugin implements the Kubernetes device plugin API
type ENIDevicePlugin struct {
	sync.RWMutex

	server *grpc.Server

	socket string
	count  int
	eniRes eniRes

	stop chan struct{}
}

func NewENIDevicePlugin(count int, eniType string) *ENIDevicePlugin {
	res, ok := eniMap[eniType]
	if !ok {
		panic("unsupported eni type " + eniType)
	}

	return &ENIDevicePlugin{
		socket: fmt.Sprintf(res.sock, time.Now().Unix()),
		count:  count,
		eniRes: res,
	}
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (m *ENIDevicePlugin) Serve() error {
	err := m.Start()
	if err != nil {
		klog.Errorf(fmt.Sprintf("can not start device plugin err: %v", err))
		return err
	}

	time.Sleep(5 * time.Second)
	klog.Infof(fmt.Sprintf("starting serve on %s", m.socket))

	err = m.Register(
		pluginapi.RegisterRequest{
			Version:      pluginapi.Version,
			Endpoint:     path.Base(m.socket),
			ResourceName: m.eniRes.resName,
		},
	)
	if err != nil {
		klog.Errorf(fmt.Sprintf("register device plugin err:%v", err))
		err = m.Stop()
		if err != nil {
			klog.Errorf(fmt.Sprintf("stop device plugin server err:%v", err))
		}
		return err
	}

	klog.Infof(fmt.Sprintf("register device plugin"))

	go m.watchKubeletRestart()

	return nil
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (m *ENIDevicePlugin) Register(request pluginapi.RegisterRequest) error {
	conn, closeConn, err := dial(pluginapi.KubeletSocket, 5*time.Second)
	if err != nil {
		return err
	}
	defer closeConn()

	client := pluginapi.NewRegistrationClient(conn)

	_, err = client.Register(context.Background(), &request)
	if err != nil {
		return err
	}
	return nil
}

// Start starts the gRPC server of the device plugin
func (m *ENIDevicePlugin) Start() error {
	if m.server != nil {
		close(m.stop)
		m.server.Stop()
	}
	err := m.cleanup()
	if err != nil {
		return err
	}

	lis, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}
	m.server = grpc.NewServer()
	pluginapi.RegisterDevicePluginServer(m.server, m)
	go func() {
		err = m.server.Serve(lis)
		if err != nil {
			klog.Errorf(fmt.Sprintf("start device plugin server err:%v", err))
		}
	}()

	// check if grpc socket server, 这里可以借鉴下
	_, closeConn, err := dial(m.socket, 5*time.Second)
	if err != nil {
		return err
	}
	closeConn()
	return nil
}

// Stop stops the gRPC server
func (m *ENIDevicePlugin) Stop() error {
	if m.server == nil {
		return nil
	}

	m.server.Stop()
	m.server = nil
	close(m.stop)

	return m.cleanup()
}

// https://github.com/kubernetes/design-proposals-archive/blob/main/resource-management/device-plugin.md#upgrading-kubelet
func (m *ENIDevicePlugin) watchKubeletRestart() {

}

func (m *ENIDevicePlugin) GetDevicePluginOptions(ctx context.Context, empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (m *ENIDevicePlugin) ListAndWatch(empty *pluginapi.Empty, server pluginapi.DevicePlugin_ListAndWatchServer) error {
	var devs []*pluginapi.Device
	for i := 0; i < m.count; i++ {
		devs = append(devs, &pluginapi.Device{ID: fmt.Sprintf("eni-%d", i), Health: pluginapi.Healthy})
	}

	err := server.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			err = server.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
			if err != nil {
				klog.Errorf(fmt.Sprintf("send device info err: %v", err))
			}
		case <-m.stop:
			return nil
		}
	}
}

// GetPreferredAllocation https://github.com/kubernetes/kubernetes/blob/v1.23.5/pkg/kubelet/cm/devicemanager/manager.go#L1065-L1089
func (m *ENIDevicePlugin) GetPreferredAllocation(ctx context.Context, request *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return nil, nil
}

func (m *ENIDevicePlugin) Allocate(ctx context.Context, request *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	response := pluginapi.AllocateResponse{
		ContainerResponses: []*pluginapi.ContainerAllocateResponse{},
	}

	for range request.GetContainerRequests() {
		response.ContainerResponses = append(response.ContainerResponses,
			&pluginapi.ContainerAllocateResponse{},
		)
	}

	return &response, nil
}

func (m *ENIDevicePlugin) PreStartContainer(ctx context.Context, request *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, func(), error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
	clientConn, err := grpc.DialContext(timeoutCtx, unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)
	if err != nil {
		cancel()
		return nil, nil, err
	}

	return clientConn, func() {
		err = clientConn.Close()
		cancel() // 不调用 cancel() 会 leads to a context leak
	}, nil
}
