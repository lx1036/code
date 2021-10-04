package v3rpc

import (
	"net"
	"testing"

	"k8s.io/klog/v2"
)

func TestGrpcWatchServer(test *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:2379") // unix or tcp, /csi/csi.sock or "127.0.0.1:2379"
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := Server()
	err = grpcServer.Serve(listener)
	if err != nil {
		klog.Fatalf("server serve fail, error:%v", err)
	}
}
