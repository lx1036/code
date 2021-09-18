package mvcc

import (
	"os"
	"testing"

	betesting "k8s-lx1036/k8s/storage/etcd/mvcc/backend/testing"

	"go.etcd.io/etcd/server/v3/lease"
)

// TestWatcherWatchID tests that each watcher provides unique watchID,
// and the watched event attaches the correct watchID.
func TestWatcherWatchID(t *testing.T) {
	b, tmpPath := betesting.NewDefaultTmpBackend(t)

	watchableStore := NewWatchableStore(b, &lease.FakeLessor{}, StoreConfig{})
	defer watchableStore.Close()
	defer watchableStore.store.Close()
	defer os.RemoveAll(tmpPath)

	watchStream := watchableStore.NewWatchStream()
	defer watchStream.Close()

	idm := make(map[WatchID]struct{})

	for i := 0; i < 10; i++ {
		id, _ := watchStream.Watch(0, []byte("foo"), nil, 0)
		if _, ok := idm[id]; ok {
			t.Errorf("#%d: id %d exists", i, id)
		}
		idm[id] = struct{}{}

		watchableStore.Put([]byte("foo"), []byte("bar"), lease.NoLease)

		resp := <-watchStream.Chan()
		if resp.WatchID != id {
			t.Errorf("#%d: watch id in event = %d, want %d", i, resp.WatchID, id)
		}

		if err := watchStream.Cancel(id); err != nil {
			t.Error(err)
		}
	}

}
