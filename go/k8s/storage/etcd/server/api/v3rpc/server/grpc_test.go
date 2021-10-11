package server

import (
	"net"
	"os"
	"testing"

	betesting "k8s-lx1036/k8s/storage/etcd/storage/backend/testing"
	"k8s-lx1036/k8s/storage/etcd/storage/mvcc"

	"go.etcd.io/etcd/server/v3/lease"
	"k8s.io/klog/v2"
)

// INFO: 不要删除，可以单独测试 server 模块
func TestGrpcWatchServer(test *testing.T) {
	b, tmpPath := betesting.NewDefaultTmpBackend()
	defer os.RemoveAll(tmpPath) // remove tmp db file
	watchableStore := mvcc.New(b, &lease.FakeLessor{}, mvcc.StoreConfig{})

	listener, err := net.Listen("tcp", "127.0.0.1:2379") // unix or tcp, /csi/csi.sock or "127.0.0.1:2379"
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()
	grpcServer := Server(watchableStore)
	err = grpcServer.Serve(listener)
	if err != nil {
		klog.Fatalf("server serve fail, error:%v", err)
	}
}
