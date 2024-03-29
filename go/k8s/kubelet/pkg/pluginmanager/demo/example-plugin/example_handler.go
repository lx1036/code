package example_plugin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/pluginmanager/demo/example-plugin/v1beta1"
	"k8s-lx1036/k8s/kubelet/pkg/pluginmanager/demo/example-plugin/v1beta2"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

type examplePluginEvent int

const (
	exampleEventValidate   examplePluginEvent = 0
	exampleEventRegister   examplePluginEvent = 1
	exampleEventDeRegister examplePluginEvent = 2
)

type exampleHandler struct {
	SupportedVersions []string
	ExpectedNames     map[string]int

	eventChans map[string]chan examplePluginEvent // map[pluginName]eventChan

	m sync.Mutex

	permitDeprecatedDir bool
}

func (p *exampleHandler) SendEvent(pluginName string, event examplePluginEvent) {
	p.eventChans[pluginName] <- event

	klog.Infof("Sending %v for plugin %s over chan %v", event, pluginName, p.eventChans[pluginName])
}

func (p *exampleHandler) EventChan(pluginName string) chan examplePluginEvent {
	return p.eventChans[pluginName]
}

func (p *exampleHandler) RegisterPlugin(pluginName, endpoint string, versions []string) error {
	klog.Infof("calling exampleHandler RegisterPlugin with %s, %s, %v", pluginName, endpoint, versions)

	//p.SendEvent(pluginName, exampleEventRegister)

	// Verifies the grpcServer is ready to serve services.
	_, conn, err := dial(endpoint, time.Second)
	if err != nil {
		return fmt.Errorf("failed dialing endpoint (%s): %v", endpoint, err)
	}
	defer conn.Close()

	// The plugin handler should be able to use any listed service API version.
	v1beta1Client := v1beta1.NewExampleClient(conn)
	v1beta2Client := v1beta2.NewExampleClient(conn)

	// Tests v1beta1 GetExampleInfo
	_, err = v1beta1Client.GetExampleInfo(context.Background(), &v1beta1.ExampleRequest{V1Beta1Field: "V1_Beta1_Field"})
	if err != nil {
		return fmt.Errorf("failed GetExampleInfo for v1beta2Client(%s): %v", endpoint, err)
	}

	// Tests v1beta1 GetExampleInfo
	_, err = v1beta2Client.GetExampleInfo(context.Background(), &v1beta2.ExampleRequest{V1Beta2Field: "V1_Beta2_Field"})
	if err != nil {
		return fmt.Errorf("failed GetExampleInfo for v1beta2Client(%s): %v", endpoint, err)
	}

	klog.Infof("called exampleHandler RegisterPlugin with %s, %s, %v", pluginName, endpoint, versions)

	return nil
}

func (p *exampleHandler) DeRegisterPlugin(pluginName string) {
	klog.Infof("calling exampleHandler DeRegisterPlugin...")

	//p.SendEvent(pluginName, exampleEventDeRegister)
}

func (p *exampleHandler) ValidatePlugin(pluginName string, endpoint string, versions []string) error {
	klog.Infof("calling exampleHandler ValidatePlugin with %s, %s, %v", pluginName, endpoint, versions)

	//p.SendEvent(pluginName, exampleEventValidate)

	/*n, ok := p.DecreasePluginCount(pluginName)
	if !ok && n > 0 {
		return fmt.Errorf("pluginName('%s') wasn't expected (count is %d)", pluginName, n)
	}*/

	if !reflect.DeepEqual(versions, p.SupportedVersions) {
		klog.Errorf("versions is not equal, %v %v", versions, p.SupportedVersions)
		return fmt.Errorf("versions('%v') != supported versions('%v')", versions, p.SupportedVersions)
	}

	// this handler expects non-empty endpoint as an example
	if len(endpoint) == 0 {
		klog.Errorf("expecting non empty endpoint")
		return errors.New("expecting non empty endpoint")
	}

	klog.Infof("called exampleHandler ValidatePlugin with %s, %s, %v", pluginName, endpoint, versions)

	return nil
}

// Dial establishes the gRPC communication with the picked up plugin socket. https://godoc.org/google.golang.org/grpc#Dial
func dial(unixSocketPath string, timeout time.Duration) (registerapi.RegistrationClient, *grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c, err := grpc.DialContext(ctx, unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial socket %s, err: %v", unixSocketPath, err)
	}

	return registerapi.NewRegistrationClient(c), c, nil
}

// NewExampleHandler provide a example handler
func NewExampleHandler(supportedVersions []string, permitDeprecatedDir bool) *exampleHandler {
	return &exampleHandler{
		SupportedVersions: supportedVersions,
		ExpectedNames:     make(map[string]int),

		eventChans:          make(map[string]chan examplePluginEvent),
		permitDeprecatedDir: permitDeprecatedDir,
	}
}
