package master

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/hashicorp/raft"
	boltdb "k8s-lx1036/k8s/storage/raft/hashicorp/bolt-store"
)

const (
	applied = "applied"
)

type Fsm struct {
	sync.Mutex

	store *boltdb.BoltStore
}

func newFsm(store *boltdb.BoltStore) *Fsm {
	fsm := &Fsm{
		store: store,
	}

	return fsm
}

func (f *Fsm) Apply(log *raft.Log) interface{} {
	var err error
	cmd := new(RaftCmd)
	if err = json.Unmarshal(log.Data, cmd); err != nil {
		return err
	}

	cmdMap := make(map[string][]byte)
	cmdMap[cmd.K] = cmd.V

	switch cmd.Op {
	case opSyncDeleteMetaNode, opSyncDeleteVol, opSyncDeleteMetaPartition, opSyncDeleteBucket, opSyncDeleteVolMountClient:
		err = f.store.BatchDelete(cmdMap)
	default:
		err = f.store.BatchPut(cmdMap)
	}

	return err
}

func (f *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	return f.store, nil
}

func (f *Fsm) Restore(snapshot io.ReadCloser) error {
	f.Lock()
	defer f.Unlock()

	dst, err := io.ReadAll(snapshot)
	if err != nil {
		return err
	}

	return f.store.Restore(dst)
}
