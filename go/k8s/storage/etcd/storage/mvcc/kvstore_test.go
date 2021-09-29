package mvcc

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"testing"

	"go.etcd.io/etcd/server/v3/lease"
	betesting "k8s-lx1036/k8s/storage/etcd/storage/backend/testing"
)

func TestStoreRev(t *testing.T) {
	b, tmpPath := betesting.NewDefaultTmpBackend(t)
	s := NewStore(b, &lease.FakeLessor{}, StoreConfig{})
	defer s.Close()
	defer os.RemoveAll(tmpPath)

	for i := 1; i <= 3; i++ {
		s.Put([]byte("foo"), []byte("bar"), lease.NoLease)
		// store current revision: 2,3,4, store启动时默认初始是 1
		if r := s.Rev(); r != int64(i+1) {
			t.Errorf("#%d: rev = %d, want %d", i, r, i+1)
		}

		result, err := s.Range(context.TODO(), []byte("foo"), nil, RangeOptions{})
		if err != nil {
			klog.Fatal(err)
		}
		klog.Infof(fmt.Sprintf("%+v", *result))
	}
}
