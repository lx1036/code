package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
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

func TestGrpcWatchClient(test *testing.T) {
	var dialOptions []grpc.DialOption
	dialOptions = append(dialOptions,
		grpc.WithInsecure(),                   // Don't use TLS, it's usually local Unix domain socket in a container.
		grpc.WithBackoffMaxDelay(time.Second), // Retry every second after failure.
		grpc.WithBlock(),                      // Block until connection succeeds.
		grpc.WithChainUnaryInterceptor(
			LogGRPC, // Log all messages.
		),
	)
	conn, err := grpc.Dial("tcp://127.0.0.1:2379", dialOptions...)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("error connecting to grpc watch server: %v", err))
	}

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

// LogGRPC is gPRC unary interceptor for logging of CSI messages at level 5. It removes any secrets from the message.
func LogGRPC(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	klog.Infof("GRPC call: %s", method)
	klog.Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	err := invoker(ctx, method, req, reply, cc, opts...)
	klog.Infof("GRPC response: %s", protosanitizer.StripSecrets(reply))
	klog.Infof("GRPC error: %v", err)
	return err
}
