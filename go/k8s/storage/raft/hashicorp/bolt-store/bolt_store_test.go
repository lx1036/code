package bolt_store

import (
	"encoding/binary"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	bolt "go.etcd.io/bbolt"
)

func TestBoltStoreImplements(test *testing.T) {
	var store interface{} = &BoltStore{}
	if _, ok := store.(raft.StableStore); !ok {
		test.Fatalf("BoltStore does not implement raft.StableStore")
	}
	if _, ok := store.(raft.LogStore); !ok {
		test.Fatalf("BoltStore does not implement raft.LogStore")
	}
}

func TestBoltOptionsTimeout(t *testing.T) {
	fh, err := ioutil.TempFile(".", "bolt")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	os.Remove(fh.Name())
	defer os.Remove(fh.Name())
	options := Options{
		Path: fh.Name(),
		BoltOptions: &bolt.Options{
			Timeout: time.Second / 10,
		},
	}
	store, err := New(options)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer store.Close()
	// trying to open it again should timeout
	doneCh := make(chan error, 1)
	go func() {
		_, err := New(options)
		doneCh <- err
	}()
	select {
	case err := <-doneCh:
		klog.Info(err.Error())
		if err == nil || err.Error() != "timeout" {
			t.Errorf("Expected timeout error but got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Gave up waiting for timeout response")
	}
}

func TestBoltOptionsReadOnly(t *testing.T) {
	// INFO: (1) 写一个 raftlog 到 boltdb
	fh, err := ioutil.TempFile(".", "bolt")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(fh.Name())
	store, err := NewBoltStore(fh.Name())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	// Create the log
	log := &raft.Log{
		Data:  []byte("log1"),
		Index: 1,
	}
	// Attempt to store the log
	if err := store.StoreLog(log); err != nil {
		t.Fatalf("err: %s", err)
	}

	store.Close()

	// INFO: (1) 从 boltdb 读取一个 raftlog
	options := Options{
		Path: fh.Name(),
		BoltOptions: &bolt.Options{
			Timeout:  time.Second / 10,
			ReadOnly: true,
		},
	}
	roStore, err := New(options)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer roStore.Close()
	result := new(raft.Log)
	if err := roStore.GetLog(1, result); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Ensure the log comes back the same
	if !reflect.DeepEqual(log, result) {
		t.Errorf("bad: %v", result)
	}
	// Attempt to store the log, should fail on a read-only store
	err = roStore.StoreLog(log)
	if err != bolt.ErrDatabaseReadOnly {
		t.Errorf("expecting error %v, but got %v", bolt.ErrDatabaseReadOnly, err)
	}
}

func TestNewBoltStore(t *testing.T) {
	fh, err := ioutil.TempFile(".", "bolt")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	os.Remove(fh.Name())

	// Successfully creates and returns a store
	store, err := NewBoltStore(fh.Name())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	// Ensure the file was created
	if store.path != fh.Name() {
		t.Fatalf("unexpected file path %q", store.path)
	}
	if _, err := os.Stat(fh.Name()); err != nil {
		t.Fatalf("err: %s", err)
	}
	// Close the store so we can open again
	if err := store.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Ensure our tables were created
	db, err := bolt.Open(fh.Name(), fs.ModePerm, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(fh.Name())
	tx, err := db.Begin(true)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if _, err := tx.CreateBucket([]byte(dbLogs)); err != bolt.ErrBucketExists {
		t.Fatalf("bad: %v", err)
	}
	if _, err := tx.CreateBucket([]byte(dbConf)); err != bolt.ErrBucketExists {
		t.Fatalf("bad: %v", err)
	}
}

func TestBoltStoreFirstIndex(t *testing.T) {
	fh, err := ioutil.TempFile(".", "bolt")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	os.Remove(fh.Name())
	// Successfully creates and returns a store
	store, err := NewBoltStore(fh.Name())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer store.Close()
	defer os.Remove(store.path)

	// Should get 0 index on empty log
	idx, err := store.FirstIndex()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 0 {
		t.Fatalf("bad: %v", idx)
	}

	// Set a mock raft log
	logs := []*raft.Log{
		{
			Index: 1,
			Data:  []byte("log1"),
		},
		{
			Index: 2,
			Data:  []byte("log2"),
		},
		{
			Index: 3,
			Data:  []byte("log3"),
		},
	}
	if err := store.StoreLogs(logs); err != nil {
		t.Fatalf("bad: %s", err)
	}

	// Fetch the first Raft index
	idx, err = store.FirstIndex()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 1 {
		t.Fatalf("bad: %d", idx)
	}
}

func TestBoltStoreLastIndex(t *testing.T) {
	fh, err := ioutil.TempFile(".", "bolt")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	os.Remove(fh.Name())
	// Successfully creates and returns a store
	store, err := NewBoltStore(fh.Name())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer store.Close()
	defer os.Remove(store.path)

	// Should get 0 index on empty log
	idx, err := store.LastIndex()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 0 {
		t.Fatalf("bad: %v", idx)
	}

	// Set a mock raft log
	logs := []*raft.Log{
		{
			Index: 1,
			Data:  []byte("log1"),
		},
		{
			Index: 2,
			Data:  []byte("log2"),
		},
		{
			Index: 3,
			Data:  []byte("log3"),
		},
	}
	if err := store.StoreLogs(logs); err != nil {
		t.Fatalf("bad: %s", err)
	}

	// Fetch the first Raft index
	idx, err = store.LastIndex()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 3 {
		t.Fatalf("bad: %d", idx)
	}
}

func TestBoltStoreGetLog(t *testing.T) {
	fh, err := ioutil.TempFile(".", "bolt")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	os.Remove(fh.Name())
	// Successfully creates and returns a store
	store, err := NewBoltStore(fh.Name())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer store.Close()
	defer os.Remove(store.path)

	// Should return an error on non-existent log
	log := new(raft.Log)
	if err := store.GetLog(1, log); err != raft.ErrLogNotFound {
		t.Fatalf("expected raft log not found error, got: %v", err)
	}

	// Set a mock raft log
	logs := []*raft.Log{
		{
			Index: 1,
			Data:  []byte("log1"),
		},
		{
			Index: 2,
			Data:  []byte("log2"),
		},
		{
			Index: 3,
			Data:  []byte("log3"),
		},
	}
	if err := store.StoreLogs(logs); err != nil {
		t.Fatalf("bad: %s", err)
	}

	// Should return the proper log
	if err := store.GetLog(2, log); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !reflect.DeepEqual(log, logs[1]) {
		t.Fatalf("bad: %#v", log)
	}
}

func TestBoltStoreDeleteRange(t *testing.T) {
	fh, err := ioutil.TempFile(".", "bolt")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	os.Remove(fh.Name())
	// Successfully creates and returns a store
	store, err := NewBoltStore(fh.Name())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer store.Close()
	defer os.Remove(store.path)

	// Set a mock raft log
	logs := []*raft.Log{
		{
			Index: 1,
			Data:  []byte("log1"),
		},
		{
			Index: 2,
			Data:  []byte("log2"),
		},
		{
			Index: 3,
			Data:  []byte("log3"),
		},
	}
	if err := store.StoreLogs(logs); err != nil {
		t.Fatalf("bad: %s", err)
	}

	// Attempt to delete a range of logs
	if err := store.DeleteRange(1, 2); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Ensure the logs were deleted
	if err := store.GetLog(1, new(raft.Log)); err != raft.ErrLogNotFound {
		t.Fatalf("should have deleted log1")
	}
	if err := store.GetLog(2, new(raft.Log)); err != raft.ErrLogNotFound {
		t.Fatalf("should have deleted log2")
	}
}

func TestRaftLogJsonMarshal(test *testing.T) {
	log := &raft.Log{
		Index:      1,
		Term:       1,
		Type:       raft.LogCommand,
		Data:       []byte("data"),
		Extensions: []byte("Extensions"),
		AppendedAt: time.Now(),
	}
	value, _ := json.Marshal(log)
	// {"Index":1,"Term":1,"Type":0,"Data":"ZGF0YQ==","Extensions":"RXh0ZW5zaW9ucw==","AppendedAt":"2022-01-13T01:46:03.19639+08:00"}
	klog.Info(string(value))

	type Person struct {
		Name []byte `json:"name"`
		City string
	}
	person := &Person{Name: []byte("name"), City: "beijing"}
	value, _ = json.Marshal(person)
	klog.Info(string(value)) // {"name":"bmFtZQ==","City":"beijing"}

	buf := make([]byte, 8)
	index := uint64(1)
	binary.BigEndian.PutUint64(buf, index)
	klog.Info(string(buf), buf)                       // 空值, [0 0 0 0 0 0 0 1]
	klog.Info(binary.BigEndian.Uint64(buf), len(buf)) // 1 8

	term, _ := strconv.ParseUint("1", 10, 64)
	klog.Info(term)
}
