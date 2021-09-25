package backend

import (
	"fmt"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"testing"

	betesting "k8s-lx1036/k8s/storage/etcd/storage/backend/testing"
)

func TestSnapshot(t *testing.T) {
	b, tmpPath := betesting.NewDefaultTmpBackend(t)
	defer b.Close()
	defer os.RemoveAll(tmpPath)

	tx := b.BatchTx()
	tx.Lock()
	tx.UnsafeCreateBucket(Test)
	tx.UnsafePut(Test, []byte("foo"), []byte("bar"))
	tx.Unlock()
	b.ForceCommit()

	// write snapshot to a new file
	f, err := ioutil.TempFile(filepath.Base("./tmp"), "test_snapshot")
	if err != nil {
		t.Fatal(err)
	}
	snapshot := b.Snapshot()
	n, err := snapshot.WriteTo(f)
	if err != nil {
		t.Fatal(err)
	}
	klog.Infof(fmt.Sprintf("%d bytes write to file %s", n, f.Name()))

	// new backend from snapshot file
	b2 := NewDefaultBackend(f.Name())
	defer b2.Close()
	tx2 := b2.BatchTx()
	tx2.RLock()
	keys, values := tx2.UnsafeRange(Test, []byte("foo"), nil, 0)
	for index, key := range keys {
		value := values[index]
		klog.Infof(fmt.Sprintf("key %s, value %s", string(key), string(value)))
	}
	if len(keys) != 1 || len(values) != 1 {
		t.Errorf("len(kvs) = %d, want 1", len(keys))
	}
	tx2.RUnlock()
}
