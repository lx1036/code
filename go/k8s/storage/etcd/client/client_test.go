package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/klog/v2"
)

const (
	EtcdDefaultRequestTimeout = 5 * time.Second
	EtcdDefaultDialTimeout    = 5 * time.Second
)

// etcd --name=etcd1 --data-dir=cluster1/etcd1 --initial-advertise-peer-urls=http://127.0.0.1:2380 --advertise-client-urls=http://127.0.0.1:2379 --listen-peer-urls=http://0.0.0.0:2380 --listen-client-urls=http://0.0.0.0:2379 --initial-cluster="etcd1=http://127.0.0.1:2380" --initial-cluster-token=abc123 --initial-cluster-state=new
// etcdctl member add etcd2 --peer-urls="http://127.0.0.1:12380"
// etcd --name=etcd2 --data-dir=cluster1/etcd2 --initial-advertise-peer-urls=http://127.0.0.1:12380 --advertise-client-urls=http://127.0.0.1:12379 --listen-peer-urls=http://0.0.0.0:12380 --listen-client-urls=http://0.0.0.0:12379 --initial-cluster="etcd1=http://127.0.0.1:2380,etcd2=http://127.0.0.1:12380" --initial-cluster-state=existing
// etcdctl member add etcd3 --peer-urls="http://127.0.0.1:22380"
// etcd --name=etcd3 --data-dir=cluster1/etcd3 --initial-advertise-peer-urls=http://127.0.0.1:22380 --advertise-client-urls=http://127.0.0.1:22379 --listen-peer-urls=http://0.0.0.0:22380 --listen-client-urls=http://0.0.0.0:22379 --initial-cluster="etcd1=http://127.0.0.1:2380,etcd2=http://127.0.0.1:12380,etcd3=http://127.0.0.1:22380" --initial-cluster-state=existing
// etcdctl member add etcd4 --peer-urls="http://127.0.0.1:32380" --learner
// etcd --name=etcd4 --data-dir=cluster1/etcd4 --initial-advertise-peer-urls=http://127.0.0.1:32380 --advertise-client-urls=http://127.0.0.1:32379 --listen-peer-urls=http://0.0.0.0:32380 --listen-client-urls=http://0.0.0.0:32379 --initial-cluster="etcd1=http://127.0.0.1:2380,etcd2=http://127.0.0.1:12380,etcd3=http://127.0.0.1:22380,etcd4=http://127.0.0.1:32380" --initial-cluster-state=existing

func TestClient(test *testing.T) {
	clientURLs := []string{"http://127.0.0.1:2379"}
	cfg := clientv3.Config{
		Endpoints:   clientURLs,
		DialTimeout: EtcdDefaultDialTimeout,
	}
	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		klog.Fatal(err)
	}
	defer etcdClient.Close()

	// INFO: 添加一个新的 member，这里重点是 --initial-cluster-state=existing
	ctx, cancel := context.WithTimeout(context.TODO(), EtcdDefaultRequestTimeout)
	defer cancel()

	memberListResponse, err := etcdClient.MemberList(ctx)
	if err != nil {
		klog.Fatal(err)
	}

	for _, member := range memberListResponse.Members {
		klog.Infof(fmt.Sprintf("member %+v", *member))
	}
}

func TestStruct(test *testing.T) {
	type Metadata struct {
		Index int
		Term  int
	}
	type Snapshot struct {
		metadata Metadata
	}
	type MemoryStorage struct {
		snapshot Snapshot
	}

	memoryStorage := &MemoryStorage{snapshot: Snapshot{metadata: Metadata{
		Index: 0,
		Term:  0,
	}}}
	memoryStorage.snapshot.metadata.Index = 1

	klog.Infof(fmt.Sprintf("%+v", *memoryStorage)) // {snapshot:{metadata:{Index:1 Term:0}}}
}
