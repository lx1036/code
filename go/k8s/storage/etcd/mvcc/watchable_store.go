package mvcc

import (
	//"fmt"
	"sync"
	"time"

	"go.etcd.io/etcd/server/v3/lease"
	"go.etcd.io/etcd/server/v3/mvcc/backend"
	//"go.etcd.io/etcd/server/v3/mvcc/buckets"
	//"k8s.io/klog/v2"
)

var (
	// chanBufLen is the length of the buffered chan
	// for sending out watched events.
	// See https://github.com/etcd-io/etcd/issues/11906 for more detail.
	chanBufLen = 128

	// maxWatchersPerSync is the number of watchers to sync in a single batch
	maxWatchersPerSync = 512
)

// cancelFunc updates unsynced and synced maps when running
// cancel operations.
type cancelFunc func()

type watchable interface {
	watch(key, end []byte, startRev int64, id WatchID, ch chan<- WatchResponse, fcs ...FilterFunc) (*watcher, cancelFunc)
}

type watchableStore struct {
	wg sync.WaitGroup
	// mu protects watcher groups and batches. It should never be locked
	// before locking store.mu to avoid deadlock.
	mu sync.RWMutex

	*store

	// INFO: contains all unsynced watchers that needs to sync with events that have happened
	unsynced watcherGroup
	// contains all synced watchers that are in sync with the progress of the store.
	// The key of the map is the key that the watcher watches on.
	synced watcherGroup

	// victims are watcher batches that were blocked on the watch channel
	//victims  []watcherBatch
	victimCh chan struct{}

	stopC chan struct{}
}

// NewWatchableStore
// INFO: 该对象是etcd最核心的一个功能，watch 功能，可以 watch key 和 watch range keys
//  会启动两个goroutine,
func NewWatchableStore(b backend.Backend, le lease.Lessor, cfg StoreConfig) *watchableStore {
	s := &watchableStore{
		store:    NewStore(b, le, cfg),
		victimCh: make(chan struct{}, 1),
		//unsynced: newWatcherGroup(),
		//synced:   newWatcherGroup(),
		stopC: make(chan struct{}),
	}
	s.store.ReadView = &readView{s}
	s.store.WriteView = &writeView{s}
	if s.le != nil {
		// use this store as the deleter so revokes trigger watch events
		s.le.SetRangeDeleter(func() lease.TxnDelete { return s.Write() })
	}

	s.wg.Add(2)
	go s.syncWatchersLoop()
	go s.syncVictimsLoop()
	return s
}

func (s *watchableStore) Close() {
	close(s.stopC)
	s.wg.Wait()
	s.store.Close()
}

// INFO: 每100ms去查下unsynced watchers，然后选择一批unsynced watchers去追赶数据，等追赶上再去放到synced watchers组里
func (s *watchableStore) syncWatchersLoop() {
	defer s.wg.Done()
	for {
		s.mu.RLock()
		st := time.Now()
		lastUnsyncedWatchers := s.unsynced.size()
		s.mu.RUnlock()

		unsyncedWatchers := 0
		if lastUnsyncedWatchers > 0 { // INFO: 还有 unsynced watchers
			unsyncedWatchers = s.syncWatchers()
		}
		syncDuration := time.Since(st)

		waitDuration := 100 * time.Millisecond
		// more work pending?
		if unsyncedWatchers != 0 && lastUnsyncedWatchers > unsyncedWatchers {
			// be fair to other store operations by yielding time taken
			waitDuration = syncDuration
		}
		select {
		case <-time.After(waitDuration):
		case <-s.stopc:
			return
		}
	}
}

// INFO: 选择一批unsynced watchers去追赶数据，等追赶上再去放到synced watchers组里, 返回unsynced watchers里还剩多少
func (s *watchableStore) syncWatchers() int {
	/*// INFO: 先加锁
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.unsynced.size() == 0 {
		return 0
	}

	// TODO: 为何revision还得加锁
	s.store.revMu.RLock()
	defer s.store.revMu.RUnlock()

	// INFO:(1)选择一批 unsynced watchers
	// in order to find key-value pairs from unsynced watchers, we need to
	// find min revision index, and these revisions can be used to
	// query the backend store of key-value pairs
	curRev := s.store.currentRev
	compactionRev := s.store.compactMainRev
	wg, minRev := s.unsynced.choose(maxWatchersPerSync, curRev, compactionRev)

	// UnsafeRange returns keys and values. And in boltdb, keys are revisions.
	// values are actual key-value pairs in backend.
	tx := s.store.b.ReadTx()
	tx.RLock()
	revs, vs := tx.UnsafeRange(buckets.Key, minBytes, maxBytes, 0)
	tx.RUnlock()
	evs := kvsToEvents(wg, revs, vs)

	var victims watcherBatch
	wb := newWatcherBatch(wg, evs)
	for w := range wg.watchers {

		watchResponse := WatchResponse{WatchID: w.id, Events: eb.evs, Revision: curRev}
		if w.send(watchResponse) {
			klog.Infof(fmt.Sprintf("[syncWatchers]fail to send watch response %+v", watchResponse))
		} else {
			if victims == nil {
				victims = make(watcherBatch)
			}
			w.victim = true
		}

	}
	s.addVictim(victims)

	return s.unsynced.size()*/

	return 0
}

func (s *watchableStore) syncVictimsLoop() {
	defer s.wg.Done()

}

// INFO: 每一个 key 都有其对应的 watcher 对象
type watcher struct {
	// the watcher key
	key []byte
	// end indicates the end of the range to watch.
	// If end is set, the watcher is on a range.
	end []byte

	// victim is set when ch is blocked and undergoing victim processing
	victim bool

	// compacted is set when the watcher is removed because of compaction
	compacted bool

	// restore is true when the watcher is being restored from leader snapshot
	// which means that this watcher has just been moved from "synced" to "unsynced"
	// watcher group, possibly with a future revision when it was first added
	// to the synced watcher
	// "unsynced" watcher revision must always be <= current revision,
	// except when the watcher were to be moved from "synced" watcher group
	restore bool

	// minRev is the minimum revision update the watcher will accept
	minRev int64
	id     WatchID

	fcs []FilterFunc
	// a chan to send out the watch response.
	// The chan might be shared with other watchers.
	ch chan<- WatchResponse
}

func (s *watchableStore) watch(key, end []byte, startRev int64, id WatchID, ch chan<- WatchResponse, fcs ...FilterFunc) (*watcher, cancelFunc) {
	w := &watcher{
		key:    key,
		end:    end,
		minRev: startRev,
		id:     id,
		ch:     ch,
		fcs:    fcs,
	}

	return w, nil
}
