package pkg

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"sync"

	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	"k8s.io/klog/v2"
)

// a key-value store backed by raft
type KVStore struct {
	proposeC    chan<- string // channel for proposing updates
	mu          sync.RWMutex
	kvStore     map[string]string // current committed key-value pairs
	snapshotter *snap.Snapshotter
}

type kv struct {
	Key string
	Val string
}

func NewKVStore(snapshotter *snap.Snapshotter, proposeC chan<- string, commitC <-chan *commit, errorC <-chan error) *KVStore {
	s := &KVStore{
		proposeC:    proposeC,
		kvStore:     make(map[string]string),
		snapshotter: snapshotter,
	}

	// read commits from raft into kvStore map until error
	go s.readCommits(commitC, errorC)

	return s
}

func (s *KVStore) readCommits(commitC <-chan *commit, errorC <-chan error) {
	for commit := range commitC { // commitC 没有值不会阻塞，而 <-commitC 会阻塞
		if commit == nil {
			continue
		}

		for _, data := range commit.data {
			var dataKv kv
			dec := gob.NewDecoder(bytes.NewBufferString(data))
			if err := dec.Decode(&dataKv); err != nil {
				klog.Fatalf("raftexample: could not decode message (%v)", err)
			}

			s.mu.Lock()
			s.kvStore[dataKv.Key] = dataKv.Val
			s.mu.Unlock()
		}
	}
}

func (s *KVStore) GetSnapshot() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return json.Marshal(s.kvStore)
}

// Propose 往 proposeC channel里写key-value，生产者
func (s *KVStore) Propose(k string, v string) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(kv{k, v}); err != nil {
		klog.Fatal(err)
	}

	s.proposeC <- buf.String()
}

//
func (s *KVStore) Lookup(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.kvStore[key]
	return v, ok
}
