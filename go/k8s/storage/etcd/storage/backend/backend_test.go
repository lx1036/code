package backend

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"

	"k8s.io/klog/v2"
)

func newTmpBackend() (*backend, string) {
	dir := "tmp"
	os.MkdirAll(dir, 0777)
	tmpPath := filepath.Join(dir, "db.txt")
	b := NewDefaultBackend(tmpPath)

	return b, tmpPath
}

func TestConcurrentReadTxn(test *testing.T) {
	b, tmpPath := newTmpBackend()
	defer b.Close()
	//defer klog.Info(tmpPath)
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

	// (1)ConcurrentReadTx
	concurrentReadTxn := b.ConcurrentReadTx()
	concurrentReadTxn.RLock() // no-op
	keys, values := concurrentReadTxn.UnsafeRange(Key, []byte("abc"), []byte("xyz"), 0)
	concurrentReadTxn.RUnlock()
	for index, key := range keys { // {"abc","def","overwrite"}, {"ABC","DEF","2"}
		klog.Infof(fmt.Sprintf("key: %s, value: %s", string(key), string(values[index])))
	}

	// (2)ReadTx
	readTxn := b.ReadTx()
	readTxn.RLock()
	keys, values = readTxn.UnsafeRange(Key, []byte("abc"), []byte("xyz"), 0)
	readTxn.RUnlock()
	for index, key := range keys { // {"abc","def","overwrite"}, {"ABC","DEF","2"}
		klog.Infof(fmt.Sprintf("key: %s, value: %s", string(key), string(values[index])))
	}

	// (3)Interval commit
	time.Sleep(time.Second)
	_ = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("key"))
		value := bucket.Get([]byte("abc"))
		if bytes.Compare(value, []byte("ABC")) != 0 {
			test.Errorf("want %s, got %s", []byte("ABC"), value)
		}
		klog.Infof(fmt.Sprintf("value: %s", value))

		return nil
	})

	// (4)backup/restore snapshot
	snapshotPath := filepath.Join("tmp", "snapshot.txt")
	file, _ := os.OpenFile(snapshotPath, os.O_CREATE|os.O_RDWR, 0777)
	defer os.RemoveAll(snapshotPath)
	klog.Infof(fmt.Sprintf("fd: %+v, file name: %s", file.Fd(), file.Name()))
	snapshot := b.Snapshot()
	num, err := snapshot.WriteTo(file)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infof(fmt.Sprintf("write %d bytes to snapshot file", num))
	b2 := NewDefaultBackend(file.Name())
	defer b2.Close()
	writeTxn3 := b2.BatchTx()
	writeTxn3.Lock()
	keys, values = writeTxn3.UnsafeRange(Key, []byte("abc"), []byte("xyz"), 0)
	writeTxn3.Unlock()
	for index, key := range keys { // {"abc","def","overwrite"}, {"ABC","DEF","2"}
		klog.Infof(fmt.Sprintf("key: %s, value: %s", string(key), string(values[index])))
	}
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

type intPairs []int

func (d intPairs) Len() int           { return len(d) }
func (d intPairs) Less(i, j int) bool { return d[i] < d[j] }
func (d intPairs) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func TestSortStable(test *testing.T) {
	a := intPairs{1, 3, 66, 4, 6, 19, 5}
	sort.Stable(a)
	klog.Info(a) // [1 3 4 5 6 19 66]
}
