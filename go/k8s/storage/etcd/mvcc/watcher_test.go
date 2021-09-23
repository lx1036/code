package mvcc

import (
	"os"
	"testing"

	betesting "k8s-lx1036/k8s/storage/etcd/mvcc/backend/testing"

	"go.etcd.io/etcd/server/v3/lease"
	"k8s.io/klog/v2"
)

// INFO: 测试 etcd watch feature
func TestSyncedWatcherGroup(t *testing.T) {
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

	// TODO: watchable_store_test.go::TestSyncWatchers()
}
