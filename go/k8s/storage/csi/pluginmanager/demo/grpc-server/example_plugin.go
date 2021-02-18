package main

import (
	"context"
	"fmt"
	"net"
	"sync"

	"k8s-lx1036/k8s/storage/csi/pluginmanager/demo/example-plugin/v1beta1"
	"k8s-lx1036/k8s/storage/csi/pluginmanager/demo/example-plugin/v1beta2"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

// examplePlugin is a sample plugin to work with plugin watcher
type examplePlugin struct {
	grpcServer         *grpc.Server
	wg                 sync.WaitGroup
	registrationStatus chan registerapi.RegistrationStatus // for testing
	endpoint           string                              // for testing
	pluginName         string
	pluginType         string
	versions           []string
}

func (e *examplePlugin) GetInfo(context context.Context, request *registerapi.InfoRequest) (*registerapi.PluginInfo, error) {
	return &registerapi.PluginInfo{
		Type:              e.pluginType,
		Name:              e.pluginName,
		Endpoint:          e.endpoint,
		SupportedVersions: e.versions,
	}, nil
}

func (e *examplePlugin) NotifyRegistrationStatus(context context.Context, status *registerapi.RegistrationStatus) (*registerapi.RegistrationStatusResponse, error) {
	klog.Infof("Registration is: %v\n", status)

	//if e.registrationStatus != nil {
	//	e.registrationStatus <- *status
	//}

	return &registerapi.RegistrationStatusResponse{}, nil
}

type pluginServiceV1Beta1 struct {
	server *examplePlugin
}

func (p *pluginServiceV1Beta1) GetExampleInfo(ctx context.Context, request *v1beta1.ExampleRequest) (*v1beta1.ExampleResponse, error) {
	klog.Infof("GetExampleInfo v1beta1_field: %s", request.V1Beta1Field)
	return &v1beta1.ExampleResponse{}, nil
}

type pluginServiceV1Beta2 struct {
	server *examplePlugin
}

func (p pluginServiceV1Beta2) GetExampleInfo(ctx context.Context, request *v1beta2.ExampleRequest) (*v1beta2.ExampleResponse, error) {
	klog.Infof("GetExampleInfo v1beta2_field: %s", request.V1Beta2Field)
	return &v1beta2.ExampleResponse{}, nil
}

// Serve starts a pluginwatcher server and one or more of the plugin services
func (e *examplePlugin) Serve(services ...string) error {
	lis, err := net.Listen("unix", e.endpoint)
	if err != nil {
		return err
	}

	klog.Infof("%s server started at: %s\n", e.pluginName, e.endpoint)
	e.grpcServer = grpc.NewServer()

	// Registers kubelet plugin watcher api.
	registerapi.RegisterRegistrationServer(e.grpcServer, e)

	for _, service := range services {
		switch service {
		case "v1beta1":
			v1beta1.RegisterExampleServer(e.grpcServer, &pluginServiceV1Beta1{})
		case "v1beta2":
			v1beta2.RegisterExampleServer(e.grpcServer, &pluginServiceV1Beta2{})
		default:
			return fmt.Errorf("unsupported service: '%s'", service)
		}
	}

	// Starts service
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		defer e.grpcServer.Stop()
		// Blocking call to accept incoming connections.
		if err := e.grpcServer.Serve(lis); err != nil {
			klog.Errorf("example server stopped serving: %v", err)
		}
	}()

	//e.wg.Wait()

	return nil
}

// NewTestExamplePlugin returns an initialized examplePlugin instance for testing
func NewTestExamplePlugin(pluginName string, pluginType string, endpoint string, advertisedVersions ...string) *examplePlugin {
	return &examplePlugin{
		pluginName:         pluginName,
		pluginType:         pluginType,
		endpoint:           endpoint,
		versions:           advertisedVersions,
		registrationStatus: make(chan registerapi.RegistrationStatus),
	}
}
