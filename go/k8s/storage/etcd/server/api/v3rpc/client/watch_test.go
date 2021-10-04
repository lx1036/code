package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"k8s-lx1036/k8s/storage/etcd/server/api/v3rpc/server"
	betesting "k8s-lx1036/k8s/storage/etcd/storage/backend/testing"
	"k8s-lx1036/k8s/storage/etcd/storage/mvcc"

	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/lease"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

func GetEtcdClient() *clientv3.Client {
	cfg := clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:12379"},
		DialTimeout: 5 * time.Second,
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		klog.Fatal(err)
	}

	return client
}

func TestWatch(test *testing.T) {
	client := GetEtcdClient()
	defer client.Close()

	revision := int64(4)
	ctx := context.TODO()
	for {
		response := client.Watch(ctx, "hello", clientv3.WithRev(revision), clientv3.WithPrefix())
		for watchResponse := range response {
			if watchResponse.CompactRevision != 0 {
				klog.Errorf(fmt.Sprintf("required revision has been compacted, use the compact revision:%d, required-revision:%d", watchResponse.CompactRevision, revision))
				revision = watchResponse.CompactRevision // 重新从 watchResponse.CompactRevision 开始 watch
				break
			}
			if watchResponse.Canceled {
				klog.Errorf(fmt.Sprintf("watcher is canceled with revision: %d error: %v", revision, watchResponse.Err()))
				return
			}
			for _, event := range watchResponse.Events {
				klog.Infof(fmt.Sprintf("Type:%s Key:%s CreateRevision:%d ModRevision:%d Version:%d Value:%s Lease:%d",
					event.Type.String(), string(event.Kv.Key), event.Kv.CreateRevision, event.Kv.ModRevision, event.Kv.Version, event.Kv.Value, event.Kv.Lease))
			}

			revision = watchResponse.Header.Revision
		}

		select {
		case <-ctx.Done():
			// server closed, return
			return
		default:
		}
	}
}

// go test -v -run ^TestGrpcWatchClient$ .
// INFO: 不要删除，可以单独测试 client 模块
func TestGrpcWatchClient(test *testing.T) {
	// launch grpc client
	var dialOptions []grpc.DialOption
	dialOptions = append(dialOptions,
		grpc.WithInsecure(), // Don't use TLS, it's usually local Unix domain socket in a container.
		grpc.WithChainUnaryInterceptor(
			LogGRPC, // Log all messages.
		),
	)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "127.0.0.1:2379", dialOptions...) // 127.0.0.1:2379
	if err != nil {
		klog.Fatalf(fmt.Sprintf("error connecting to grpc watch server: %v", err))
	}
	defer conn.Close()
	watcher := NewWatcher(conn)
	response := watcher.Watch(context.TODO(), "hello")
	for watchResponse := range response {
		klog.Infof(fmt.Sprintf("ResponseHeader: %+v", watchResponse.Header))
		for _, event := range watchResponse.Events {
			klog.Infof(fmt.Sprintf("Type: %s, KV: Key(%s) CreateRevision(%d) ModRevision(%d) Version(%d) Value(%s), PreKV: Key(%s) CreateRevision(%d) ModRevision(%d) Version(%d) Value(%s)",
				event.Type.String(),
				string(event.Kv.Key), event.Kv.CreateRevision, event.Kv.ModRevision, event.Kv.Version, string(event.Kv.Value),
				string(event.PrevKv.Key), event.PrevKv.CreateRevision, event.PrevKv.ModRevision, event.PrevKv.Version, string(event.PrevKv.Value),
			))
		}
	}
}

// go test -v -run ^TestGrpcWatch$ .
func TestGrpcWatch(test *testing.T) {
	b, tmpPath := betesting.NewDefaultTmpBackend()
	defer os.RemoveAll(tmpPath) // remove tmp db file
	watchableStore := mvcc.New(b, &lease.FakeLessor{}, mvcc.StoreConfig{})

	listener, err := net.Listen("tcp", "127.0.0.1:2379") // unix or tcp, /csi/csi.sock or "127.0.0.1:2379"
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	go func() { // launch grpc server
		grpcServer := server.Server(watchableStore)
		err = grpcServer.Serve(listener)
		if err != nil {
			klog.Fatalf("server serve fail, error:%v", err)
		}
	}()

	// INFO: 每 5min 写事务
	index := 0
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		watchableStore.Put([]byte("hello"), []byte(fmt.Sprintf("world-%d", index)), lease.NoLease)
		index++
	}, time.Second*5)

	time.Sleep(time.Second)

	// launch grpc client
	var dialOptions []grpc.DialOption
	dialOptions = append(dialOptions,
		grpc.WithInsecure(), // Don't use TLS, it's usually local Unix domain socket in a container.
		grpc.WithChainUnaryInterceptor(
			LogGRPC, // Log all messages.
		),
	)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	conn, err := grpc.DialContext(ctx, listener.Addr().String(), dialOptions...) // 127.0.0.1:12379
	if err != nil {
		klog.Fatalf(fmt.Sprintf("error connecting to grpc watch server: %v", err))
	}
	defer conn.Close()
	watcher := NewWatcher(conn)
	response := watcher.Watch(context.TODO(), "hello")
	shutdownHandler := make(chan os.Signal, 2)
	signal.Notify(shutdownHandler, []os.Signal{os.Interrupt, syscall.SIGTERM}...)
	for {
		select {
		case watchResponse := <-response:
			klog.Infof(fmt.Sprintf("ResponseHeader: %+v", watchResponse.Header))
			for _, event := range watchResponse.Events {
				// {Type: PUT, KV: Key(hello) CreateRevision(2) ModRevision(7) Version(6) Value(world)}
				klog.Infof(fmt.Sprintf("{Type: %s, KV: Key(%s) CreateRevision(%d) ModRevision(%d) Version(%d) Value(%s)}",
					event.Type.String(),
					string(event.Kv.Key), event.Kv.CreateRevision, event.Kv.ModRevision, event.Kv.Version, string(event.Kv.Value),
				))
				if event.PrevKv != nil {
					klog.Infof(fmt.Sprintf("{Type: %s, KV: Key(%s) CreateRevision(%d) ModRevision(%d) Version(%d) Value(%s)}",
						event.Type.String(),
						string(event.PrevKv.Key), event.PrevKv.CreateRevision, event.PrevKv.ModRevision, event.PrevKv.Version, string(event.PrevKv.Value),
					))
				}
			}
		case <-shutdownHandler: // INFO: 断点断掉后会 remove tmp db file
			return
		}
	}
}

// LogGRPC is gPRC unary interceptor for logging of CSI messages at level 5. It removes any secrets from the message.
func LogGRPC(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	klog.Infof("GRPC call: %s", method)
	klog.Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	err := invoker(ctx, method, req, reply, cc, opts...)
	klog.Infof("GRPC response: %s", protosanitizer.StripSecrets(reply))
	klog.Infof("GRPC error: %v", err)
	return err
}
