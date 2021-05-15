package remote

import (
	"fmt"
	"os"

	"k8s-lx1036/k8s/kubelet/pkg/dockershim"
	"k8s-lx1036/k8s/kubelet/pkg/util"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

// maxMsgSize use 16MB as the default message size limit.
// grpc library default is 4MB
const maxMsgSize = 1024 * 1024 * 16

// DockerServer is the grpc server of dockershim.
type DockerServer struct {
	// endpoint is the endpoint to serve on.
	endpoint string
	// service is the docker service which implements runtime and image services.
	service dockershim.CRIService
	// server is the grpc server.
	server *grpc.Server
}

// NewDockerServer creates the dockershim grpc server.
func NewDockerServer(endpoint string, s dockershim.CRIService) *DockerServer {
	return &DockerServer{
		endpoint: endpoint,
		service:  s,
	}
}

// Start starts the dockershim grpc server.
func (s *DockerServer) Start() error {
	// Start the internal service.
	if err := s.service.Start(); err != nil {
		klog.ErrorS(err, "Unable to start docker service")
		return err
	}

	klog.V(2).InfoS("Start dockershim grpc server")
	l, err := util.CreateListener(s.endpoint)
	if err != nil {
		return fmt.Errorf("failed to listen on %q: %v", s.endpoint, err)
	}
	// Create the grpc server and register runtime and image services.
	s.server = grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	)
	runtimeapi.RegisterRuntimeServiceServer(s.server, s.service)
	runtimeapi.RegisterImageServiceServer(s.server, s.service)
	go func() {
		if err := s.server.Serve(l); err != nil {
			klog.ErrorS(err, "Failed to serve connections")
			os.Exit(1)
		}
	}()

	return nil
}
