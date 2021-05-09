package etcd3

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"k8s-lx1036/k8s/apiserver/pkg/apis/example"
	examplev1 "k8s-lx1036/k8s/apiserver/pkg/apis/example/v1"
	"k8s-lx1036/k8s/apiserver/pkg/storage/value"

	"go.etcd.io/etcd/integration"
	"k8s.io/apimachinery/pkg/api/apitesting"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const defaultTestPrefix = "test!"

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

// prefixTransformer adds and verifies that all data has the correct prefix on its way in and out.
type prefixTransformer struct {
	prefix []byte
	stale  bool
	err    error
	reads  uint64
}

func (p prefixTransformer) TransformFromStorage(data []byte, ctx value.Context) (out []byte, stale bool, err error) {
	atomic.AddUint64(&p.reads, 1)
	if ctx == nil {
		panic("no context provided")
	}
	if !bytes.HasPrefix(data, p.prefix) {
		return nil, false, fmt.Errorf("value does not have expected prefix %q: %s", p.prefix, string(data))
	}
	return bytes.TrimPrefix(data, p.prefix), p.stale, p.err
}

func (p prefixTransformer) TransformToStorage(data []byte, ctx value.Context) (out []byte, err error) {
	if ctx == nil {
		panic("no context provided")
	}
	if len(data) > 0 {
		return append(append([]byte{}, p.prefix...), data...), p.err
	}
	return data, p.err
}

func init() {
	metav1.AddToGroupVersion(scheme, metav1.SchemeGroupVersion)
	utilruntime.Must(examplev1.AddToScheme(scheme))
}

func TestCreate(test *testing.T) {
	ctx, s, cluster := testSetup(test)
	defer cluster.Terminate(test)
	etcdClient := cluster.RandClient()

	key := "/testkey"
	out := &example.Pod{}
	obj := &example.Pod{ObjectMeta: metav1.ObjectMeta{Name: "foo", SelfLink: "testlink"}}

	// verify that kv pair is empty before set
	getResp, err := etcdClient.KV.Get(ctx, key)
	if err != nil {
		test.Fatalf("etcdClient.KV.Get failed: %v", err)
	}
	if len(getResp.Kvs) != 0 {
		test.Fatalf("expecting empty result on key: %s", key)
	}

	err = s.Create(ctx, key, obj, out, 0)
	if err != nil {
		test.Fatalf("Set failed: %v", err)
	}
	// basic tests of the output
	if obj.ObjectMeta.Name != out.ObjectMeta.Name {
		test.Errorf("pod name want=%s, get=%s", obj.ObjectMeta.Name, out.ObjectMeta.Name)
	}
	if out.ResourceVersion == "" {
		test.Errorf("output should have non-empty resource version")
	}
	if out.SelfLink != "" {
		test.Errorf("output should have empty self link")
	}

	//checkStorageInvariants(ctx, t, etcdClient, store, key)
}

func testSetup(t *testing.T) (context.Context, *store, *integration.ClusterV3) {
	codec := apitesting.TestCodec(codecs, examplev1.SchemeGroupVersion)
	cluster := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	// As 30s is the default timeout for testing in glboal configuration,
	// we cannot wait longer than that in a single time: change it to 10
	// for testing purposes. See apimachinery/pkg/util/wait/wait.go
	s := newStore(cluster.RandClient(), true, codec, "",
		&prefixTransformer{prefix: []byte(defaultTestPrefix)}, LeaseManagerConfig{
			ReuseDurationSeconds: 1,
			MaxObjectCount:       defaultLeaseMaxObjectCount,
		})
	ctx := context.Background()
	return ctx, s, cluster
}
