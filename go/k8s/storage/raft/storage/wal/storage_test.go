package wal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/tiglabs/raft/proto"
	"github.com/tiglabs/raft/storage/wal"
	"k8s.io/klog/v2"
)

// INFO: https://github.com/tiglabs/raft/blob/master/storage/wal/bench/main.go

func TestStorage(test *testing.T) {
	dir, err := ioutil.TempDir(".", "db_bench_")
	if err != nil {
		klog.Fatal(err)
	}
	//defer os.RemoveAll(dir)

	storage, err := wal.NewStorage(dir, nil)
	if err != nil {
		klog.Fatal(err)
	}

	type Cmd struct {
		Op    string `json:"op"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	data1, _ := json.Marshal(Cmd{
		Op:    "add",
		Key:   "hello",
		Value: "world",
	})
	data2, _ := json.Marshal(Cmd{
		Op:    "add",
		Key:   "hello1",
		Value: "world1",
	})
	entries := []*proto.Entry{
		{
			Index: 1,
			Type:  proto.EntryNormal,
			Term:  1,
			Data:  []byte("hello"),
		},
		{
			Index: 2,
			Type:  proto.EntryNormal,
			Term:  1,
			Data:  data1,
		},
		{
			Index: 3,
			Type:  proto.EntryNormal,
			Term:  1,
			Data:  data2,
		},
	}
	err = storage.StoreEntries(entries)
	if err != nil {
		klog.Fatal(err)
	}

	// INFO: fetch entry
	entries, isCompact, err := storage.Entries(1, 3, 2000) // [1,3)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Info(isCompact, len(entries)) // false 2
	for _, entry := range entries {
		// entry: &{Type:EntryNormal Term:1 Index:1 Data:[104 101 108 108 111]}, hello
		// entry: &{Type:EntryNormal Term:1 Index:2 Data:[123 34 111 ... 100 34 125]}, {"op":"add","key":"hello","value":"world"}
		klog.Infof(fmt.Sprintf("entry: %+v, %s", entry, entry.Data))
	}

	// initial state
	hardState, _ := storage.InitialState()
	klog.Infof(fmt.Sprintf("InitialState %+v", hardState)) // {Term:0 Commit:0 Vote:0}

	// term
	term, isCompact, err := storage.Term(1)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Info(term, isCompact) // 1 false

	// index
	firstIndex, _ := storage.FirstIndex()
	lastIndex, _ := storage.LastIndex()
	klog.Info(firstIndex, lastIndex) // 1 3

	// INFO: truncate
	_ = storage.Truncate(1)
	firstIndex, _ = storage.FirstIndex()
	lastIndex, _ = storage.LastIndex()
	klog.Info(firstIndex, lastIndex) // 2 3
}

func TestRaftStorage(test *testing.T) {

}
