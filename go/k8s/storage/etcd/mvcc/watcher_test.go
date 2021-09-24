package mvcc

import (
	"bytes"
	"os"
	"testing"

	betesting "k8s-lx1036/k8s/storage/etcd/mvcc/backend/testing"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/server/v3/lease"
	"k8s.io/klog/v2"
)

// INFO: 测试 etcd watch feature
func TestWatcherStore(t *testing.T) {
	b, tmpPath := betesting.NewDefaultTmpBackend(t)
	watchableStore := NewWatchableStore(b, &lease.FakeLessor{}, StoreConfig{})
	defer watchableStore.store.Close()
	defer os.RemoveAll(tmpPath)
	watchStream := watchableStore.NewWatchStream()
	defer watchStream.Close()

	// INFO: (1)测试synced 和 Cancel()
	// 写数据到boltdb中
	testKey := []byte("foo")
	testValue := []byte("bar")
	watchableStore.Put(testKey, testValue, lease.NoLease)
	watchID, err := watchStream.Watch(0, testKey, nil, 0)
	if err != nil {
		klog.Fatal(err)
	}
	if !watchableStore.synced.contains(string(testKey)) {
		t.Errorf("the key %s must be in synced watcher group", testKey)
	}
	err = watchStream.Cancel(watchID)
	if err != nil {
		klog.Fatal(err)
	}
	if watchableStore.synced.contains(string(testKey)) {
		t.Errorf("the key %s should not be in synced watcher group", testKey)
	}

	// INFO: (2)测试 Chan() 获取数据
	idm := make(map[WatchID]struct{})
	testKey1 := []byte("foo1")
	testValue2 := []byte("bar1")
	for i := 0; i < 10; i++ {
		id, _ := watchStream.Watch(0, testKey1, nil, 0)
		if _, ok := idm[id]; ok {
			t.Errorf("#%d: id %d exists", i, id)
		}
		idm[id] = struct{}{}
		watchableStore.Put(testKey1, testValue2, lease.NoLease)
		resp := <-watchStream.Chan()
		if resp.WatchID != id {
			t.Errorf("#%d: watch id in event = %d, want %d", i, resp.WatchID, id)
		}
		if err := watchStream.Cancel(id); err != nil {
			t.Error(err)
		}
		// Cancel(id) 会从 unsynced watcherGroup 中删除
		if size := watchableStore.unsynced.size(); size != 0 {
			t.Errorf("unsynced size = %d, want 0", size)
		}
	}
}

func TestSyncWatchers(t *testing.T) {
	b, tmpPath := betesting.NewDefaultTmpBackend(t)
	store := &watchableStore{
		store:    NewStore(b, &lease.FakeLessor{}, StoreConfig{}),
		unsynced: newWatcherGroup(),
		synced:   newWatcherGroup(),
	}
	defer store.store.Close()
	defer os.RemoveAll(tmpPath)
	stream := store.NewWatchStream()
	defer stream.Close()

	testKey := []byte("foo")
	testValue := []byte("bar")
	store.Put(testKey, testValue, lease.NoLease)
	watcherN := 100
	for i := 0; i < watcherN; i++ {
		// startRev=1 使得 watchers 都在 unsynced watcherGroup
		stream.Watch(0, testKey, nil, 1)
	}
	syncedWatcherSet := store.synced.watcherSetByKey(string(testKey))
	unsyncedWatcherSet := store.unsynced.watcherSetByKey(string(testKey))
	if len(syncedWatcherSet) != 0 {
		t.Fatalf("synced[string(testKey)] size = %d, want 0", len(syncedWatcherSet))
	}
	if len(unsyncedWatcherSet) != watcherN {
		t.Fatalf("unsynced size = %d, want %d", len(unsyncedWatcherSet), watcherN)
	}

	// move unsynced to synced
	store.syncWatchers()

	// INFO: 比对 synced,
	syncedWatcherSet = store.synced.watcherSetByKey(string(testKey))
	unsyncedWatcherSet = store.unsynced.watcherSetByKey(string(testKey))
	if len(syncedWatcherSet) != watcherN {
		t.Fatalf("synced[string(testKey)] size = %d, want %d", len(syncedWatcherSet), watcherN)
	}
	if len(unsyncedWatcherSet) != 0 {
		t.Fatalf("unsynced size = %d, want 0", len(unsyncedWatcherSet))
	}
	for watcher := range syncedWatcherSet {
		if watcher.minRev != store.Rev()+1 {
			t.Errorf("w.minRev = %d, want %d", watcher.minRev, store.Rev()+1)
		}
	}
	if len(stream.(*watchStream).ch) != watcherN { // ch length 必须是 100
		t.Errorf("watched event size = %d, want %d", len(stream.(*watchStream).ch), watcherN)
	}
	evs := (<-stream.(*watchStream).ch).Events
	if len(evs) != 1 {
		t.Errorf("len(evs) got = %d, want = 1", len(evs))
	}
	if evs[0].Type != mvccpb.PUT {
		t.Errorf("got = %v, want = %v", evs[0].Type, mvccpb.PUT)
	}
	if !bytes.Equal(evs[0].Kv.Key, testKey) {
		t.Errorf("got = %s, want = %s", evs[0].Kv.Key, testKey)
	}
	if !bytes.Equal(evs[0].Kv.Value, testValue) {
		t.Errorf("got = %s, want = %s", evs[0].Kv.Value, testValue)
	}
}
