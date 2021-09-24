package demo

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
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
