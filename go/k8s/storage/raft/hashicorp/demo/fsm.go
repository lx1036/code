package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/hashicorp/raft"
	"k8s.io/klog/v2"
)

type Fsm struct {
	sync.Mutex

	kvstore *KVStore
}

func (f *Fsm) Apply(log *raft.Log) interface{} {
	klog.Infof(fmt.Sprintf("apply data: %s", log.Data))

	data := strings.Split(string(log.Data), ",")
	op := data[0]
	f.Lock()
	if op == "set" {
		key := data[1]
		value := data[2]
		f.kvstore.Set(key, value)
	}
	f.Unlock()

	return nil
}

func (f *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	return f.kvstore, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
func (f *Fsm) Restore(reader io.ReadCloser) error {
	f.Lock()
	defer f.Unlock()

	dst, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return f.kvstore.Restore(dst)
}

type KVStore struct {
	sync.Mutex

	Data map[string]string
}

func NewKVStore() *KVStore {
	return &KVStore{
		Data: make(map[string]string),
	}
}

func (store *KVStore) Set(key, value string) {
	store.Lock()
	defer store.Unlock()
	store.Data[key] = value
}

func (store *KVStore) Get(key string) string {
	store.Lock()
	defer store.Unlock()

	return store.Data[key]
}

// Restore INFO: restore data from snapshot @see https://github.com/hashicorp/raft/blob/v1.3.3/fsm.go#L234-L247
func (store *KVStore) Restore(data []byte) error {
	store.Lock()
	defer store.Unlock()

	// {"hello":"world","hello1":"world1","hello2":"world2","hello3":"world3","hello4":"world4"}
	klog.Info(string(data))
	kvdata := make(map[string]string)
	err := json.Unmarshal(data, &kvdata)
	if err != nil {
		return err
	}

	for key, value := range kvdata {
		store.Data[key] = value
	}
	return nil
}

// Persist INFO: snapshot fsm data into sink @see https://github.com/hashicorp/raft/blob/v1.3.3/snapshot.go#L185-L190
func (store *KVStore) Persist(sink raft.SnapshotSink) error {
	store.Lock()
	data, _ := json.Marshal(store.Data)
	store.Unlock()

	sink.Write(data)
	return sink.Close()
}

func (store *KVStore) Release() {

}
