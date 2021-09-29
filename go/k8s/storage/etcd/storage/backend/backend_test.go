package backend

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/klog/v2"
)

func newTmpBackend() (*Backend, string) {
	dir := "tmp"
	os.MkdirAll(dir, 0777)
	tmpPath := filepath.Join(dir, "db.txt")
	b := NewDefaultBackend(tmpPath)

	return b, tmpPath
}

func TestConcurrentReadTxn(test *testing.T) {
	b, tmpPath := newTmpBackend()
	defer b.Close()
	defer os.RemoveAll(tmpPath)

	writeTxn1 := b.BatchTx()
	writeTxn1.Lock()
	writeTxn1.UnsafeCreateBucket(Key)
	writeTxn1.UnsafePut(Key, []byte("abc"), []byte("ABC"))
	writeTxn1.UnsafePut(Key, []byte("overwrite"), []byte("1"))
	writeTxn1.Unlock()

	writeTxn2 := b.BatchTx()
	writeTxn2.Lock()
	writeTxn2.UnsafePut(Key, []byte("def"), []byte("DEF"))
	writeTxn2.UnsafePut(Key, []byte("overwrite"), []byte("2"))
	writeTxn2.Unlock()

	rtx := b.ConcurrentReadTx()
	rtx.RLock() // no-op
	keys, values := rtx.UnsafeRange(Key, []byte("abc"), []byte("\xff"), 0)
	rtx.RUnlock()

	klog.Infof(fmt.Sprintf("keys: %+v, values: %+v", keys, values))
}

func TestSnapshot(t *testing.T) {
	b, tmpPath := newTmpBackend()
	defer b.Close()
	defer os.RemoveAll(tmpPath)

	tx := b.BatchTx()
	tx.Lock()
	tx.UnsafeCreateBucket(Test) // 在 boltdb 中创建 bucket
	tx.UnsafePut(Test, []byte("foo"), []byte("bar"))
	tx.Unlock()
	b.ForceCommit()

	// write snapshot to a new file
	f, err := ioutil.TempFile("tmp", "snapshot")
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
