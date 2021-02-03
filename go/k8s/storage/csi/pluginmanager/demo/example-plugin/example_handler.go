package example_plugin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	"google.golang.org/grpc"
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

func (p *exampleHandler) ValidatePlugin(pluginName string, endpoint string, versions []string) error {
	p.SendEvent(pluginName, exampleEventValidate)

	n, ok := p.DecreasePluginCount(pluginName)
	if !ok && n > 0 {
		return fmt.Errorf("pluginName('%s') wasn't expected (count is %d)", pluginName, n)
	}

	if !reflect.DeepEqual(versions, p.SupportedVersions) {
		return fmt.Errorf("versions('%v') != supported versions('%v')", versions, p.SupportedVersions)
	}

	// this handler expects non-empty endpoint as an example
	if len(endpoint) == 0 {
		return errors.New("expecting non empty endpoint")
	}

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
