package mvcc

import (
	"fmt"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/etcd/storage/backend"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/server/v3/lease"
	"k8s.io/klog/v2"
)

var (
	// chanBufLen is the length of the buffered chan
	// for sending out watched events.
	// See https://github.com/etcd-io/etcd/issues/11906 for more detail.
	chanBufLen = 128

	// maxWatchersPerSync is the number of watchers to sync in a single batch
	maxWatchersPerSync = 512
)

type WatchableKV interface {
	KV

	Watchable
}

type Watchable interface {
	// NewWatchStream returns a WatchStream that can be used to
	// watch events happened or happening on the KV.
	NewWatchStream() WatchStream
}

// INFO: watchable 只有写事务
type watchableStoreTxnWrite struct {
	TxnWrite
	store *watchableStore
}

func (s *watchableStore) Write() TxnWrite {
	return &watchableStoreTxnWrite{s.store.Write(), s}
}

func (tw *watchableStoreTxnWrite) End() {
	changes := tw.Changes()
	if len(changes) == 0 {
		tw.TxnWrite.End()
		return
	}

	rev := tw.Rev() + 1
	events := make([]mvccpb.Event, len(changes))
	for i, change := range changes {
		events[i].Kv = &changes[i]
		if change.CreateRevision == 0 {
			// INFO: 如果是 DELETE，更新 ModRevision
			events[i].Type = mvccpb.DELETE
			events[i].Kv.ModRevision = rev
		} else {
			events[i].Type = mvccpb.PUT
		}
	}

	// INFO: watch 核心功能，会在每次 put 之后再回调 notify，去 send WatchResponse 给 client
	tw.store.mu.Lock()
	tw.store.notify(rev, events)
	tw.TxnWrite.End()
	tw.store.mu.Unlock()
}

// cancelFunc updates unsynced and synced maps when running
// cancel operations.
type cancelFunc func()

type watchable interface {
	watch(key, end []byte, startRev int64, id WatchID, ch chan<- WatchResponse, fcs ...FilterFunc) (*watcher, cancelFunc)

	rev() int64
}

type watchableStore struct {
	wg sync.WaitGroup
	// mu protects watcher groups and batches. It should never be locked
	// before locking store.mu to avoid deadlock.
	mu sync.RWMutex

	*store

	// INFO: contains all unsynced watchers that needs to sync with events that have happened
	unsynced watcherGroup
	// INFO: key 是需要 watcher watch 的 key, unsynced/synced 互相转换, synced 表示已经赶上了 key 的进度
	synced watcherGroup

	// victims are watcher batches that were blocked on the watch channel
	victims  []watcherBatch
	victimCh chan struct{}

	stopC chan struct{}
}

func New(b backend.Backend, le lease.Lessor, cfg StoreConfig) WatchableKV {
	return NewWatchableStore(b, le, cfg)
}

// NewWatchableStore
// INFO: 该对象是etcd最核心的一个功能，watch 功能，可以 watch key 和 watch range keys
//  会启动两个goroutine, syncedWatchers/unsyncedWatchers/
func NewWatchableStore(b backend.Backend, le lease.Lessor, cfg StoreConfig) *watchableStore {
	s := &watchableStore{
		store:    NewStore(b, le, cfg),
		victimCh: make(chan struct{}, 1),
		unsynced: newWatcherGroup(),
		synced:   newWatcherGroup(),
		stopC:    make(chan struct{}),
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
	// INFO: 先加锁
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
	// INFO: 每次最多只能choose 512个unsynced watchers
	wg, minRev := s.unsynced.choose(maxWatchersPerSync, curRev, compactionRev)
	minBytes, maxBytes := newRevBytes(), newRevBytes()
	revToBytes(revision{main: minRev}, minBytes)
	revToBytes(revision{main: curRev + 1}, maxBytes)

	// UnsafeRange returns keys and values. And in boltdb, keys are revisions.
	// values are actual key-value pairs in backend.
	// INFO: 从boltdb中根据revisions来range search出其values
	tx := s.store.b.ReadTx()
	tx.RLock()
	revs, vs := tx.UnsafeRange(backend.Key, minBytes, maxBytes, 0)
	tx.RUnlock()
	evs := kvsToEvents(wg, revs, vs)

	victims := make(watcherBatch)
	wb := newWatcherBatch(wg, evs)
	for w := range wg.watchers {
		w.minRev = curRev + 1

		eb, ok := wb[w]
		if !ok {
			// bring un-notified watcher to synced
			s.synced.add(w)
			s.unsynced.delete(w)
			continue
		}

		if eb.moreRev != 0 {
			w.minRev = eb.moreRev
		}

		// INFO: watcher.ch <- watchResponse
		watchResponse := WatchResponse{WatchID: w.id, Events: eb.evs, Revision: curRev}
		if w.send(watchResponse) { // 成功发送
			klog.Infof(fmt.Sprintf("[syncWatchers]successfully send watch response %+v", watchResponse))
		} else {
			w.victim = true
		}

		if w.victim {
			victims[w] = eb
		} else {
			if eb.moreRev != 0 {
				// stay unsynced; more to read
				continue
			}
			s.synced.add(w)
		}
		s.unsynced.delete(w)
	}
	s.addVictim(victims)

	return s.unsynced.size()
}

// INFO: syncVictimsLoop 会监听 s.victimCh 这个 channel
func (s *watchableStore) addVictim(victim watcherBatch) {
	if victim == nil || len(victim) == 0 {
		return
	}
	s.victims = append(s.victims, victim)
	select {
	case s.victimCh <- struct{}{}:
	default:
	}
}

// kvsToEvents gets all events for the watchers from all key-value pairs
func kvsToEvents(wg *watcherGroup, revs, vals [][]byte) (evs []mvccpb.Event) {
	for i, v := range vals {
		var kv mvccpb.KeyValue
		if err := kv.Unmarshal(v); err != nil {
			klog.Errorf(fmt.Sprintf("[kvsToEvents]failed to unmarshal mvccpb.KeyValue %v", err))
			continue
		}

		if !wg.contains(string(kv.Key)) {
			continue
		}

		ty := mvccpb.PUT
		if isTombstone(revs[i]) {
			ty = mvccpb.DELETE
			// patch in mod revision so watchers won't skip
			kv.ModRevision = bytesToRev(revs[i]).main
		}
		evs = append(evs, mvccpb.Event{Kv: &kv, Type: ty})
	}
	return evs
}

// INFO:
func (s *watchableStore) syncVictimsLoop() {
	defer s.wg.Done()

	for {
		// INFO:
		for s.moveVictims() != 0 {
			// try to update all victim watchers
		}

		s.mu.RLock()
		isEmpty := len(s.victims) == 0
		s.mu.RUnlock()

		var tickC <-chan time.Time
		if !isEmpty {
			tickC = time.After(10 * time.Millisecond)
		}

		select {
		case <-tickC:
		case <-s.victimCh:
		case <-s.stopc:
			return
		}
	}
}

// moveVictims tries to update watches with already pending event data
// INFO:(2)异常场景重试机制
func (s *watchableStore) moveVictims() (moved int) {
	s.mu.Lock()
	victims := s.victims
	s.victims = nil
	s.mu.Unlock()

	var newVictim watcherBatch
	for _, wb := range victims {
		// try to send responses again
		for w, eb := range wb {
			// watcher has observed the store up to, but not including, w.minRev
			rev := w.minRev - 1
			if w.send(WatchResponse{WatchID: w.id, Events: eb.evs, Revision: rev}) {
				//pendingEventsGauge.Add(float64(len(eb.evs)))
			} else {
				if newVictim == nil {
					newVictim = make(watcherBatch)
				}
				newVictim[w] = eb
				continue
			}
			moved++
		}

		// assign completed victim watchers to unsync/sync
		s.mu.Lock()
		s.store.revMu.RLock()
		curRev := s.store.currentRev
		for w, eb := range wb {
			if newVictim != nil && newVictim[w] != nil {
				// couldn't send watch response; stays victim
				continue
			}
			w.victim = false
			if eb.moreRev != 0 {
				w.minRev = eb.moreRev
			}
			if w.minRev <= curRev {
				s.unsynced.add(w)
			} else {
				//slowWatcherGauge.Dec()
				s.synced.add(w)
			}
		}
		s.store.revMu.RUnlock()
		s.mu.Unlock()
	}

	if len(newVictim) > 0 {
		s.mu.Lock()
		s.victims = append(s.victims, newVictim)
		s.mu.Unlock()
	}

	return moved
}

// INFO: watch 核心功能，会在每次 put 之后再回调 notify，去 send WatchResponse 给 client; 把 slow watcher 放到 victim 里
func (s *watchableStore) notify(rev int64, evs []mvccpb.Event) {
	victim := make(watcherBatch)
	watcherBatch := newWatcherBatch(&s.synced, evs)
	for watcher, eventBatch := range watcherBatch {
		if eventBatch.revs != 1 {
			klog.Errorf(fmt.Sprintf("[watchableStore notify]unexpected multiple revisions in watch notification: %d", eventBatch.revs))
			continue
		}

		// INFO: 这里 watcher.ch <- WatchResponse, 重点!!!
		if watcher.send(WatchResponse{
			WatchID:  watcher.id,
			Events:   eventBatch.evs,
			Revision: rev,
		}) {
			// metrics
		} else {
			// move slow watcher to victims
			watcher.minRev = rev + 1
			watcher.victim = true
			victim[watcher] = eventBatch
			s.synced.delete(watcher)
		}
	}

	s.addVictim(victim)
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

// INFO: 把 watcher 放入 synced/unsynced watcherGroup 里, 根据当前 etcd node currentRev 来判断 synced/unsynced
func (s *watchableStore) watch(key, end []byte, startRev int64, id WatchID, ch chan<- WatchResponse, fcs ...FilterFunc) (*watcher, cancelFunc) {
	w := &watcher{
		key:    key,
		end:    end,
		minRev: startRev,
		id:     id,
		ch:     ch,
		fcs:    fcs,
	}

	s.mu.Lock()
	s.revMu.RLock()
	// INFO: 根据当前集群版本号 etcd node currentRev 来判断 synced/unsynced
	synced := startRev > s.store.currentRev || startRev == 0
	if synced {
		w.minRev = s.store.currentRev + 1
		if startRev > w.minRev {
			w.minRev = startRev
		}
		s.synced.add(w)
	} else {
		//slowWatcherGauge.Inc()
		s.unsynced.add(w)
	}
	s.revMu.RUnlock()
	s.mu.Unlock()

	return w, func() { s.cancelWatcher(w) }
}

// cancelWatcher removes references of the watcher from the watchableStore
// INFO: 从 unsynced/synced/victims 中删除 watcher
func (s *watchableStore) cancelWatcher(w *watcher) {
	for {
		s.mu.Lock()
		if s.unsynced.delete(w) {
			break
		} else if s.synced.delete(w) {
			break
		} else if w.compacted {
			break
		} else if w.ch == nil {
			// already canceled (e.g., cancel/close race)
			break
		}

		if !w.victim {
			s.mu.Unlock()
			panic("watcher not victim but not in watch groups")
		}

		var victimBatch watcherBatch
		for _, wb := range s.victims {
			if wb[w] != nil {
				victimBatch = wb
				break
			}
		}
		if victimBatch != nil {
			delete(victimBatch, w)
			break
		}

		// victim being processed so not accessible; retry
		s.mu.Unlock()
		time.Sleep(time.Millisecond)
	}

	w.ch = nil
	s.mu.Unlock()
}

func (s *watchableStore) rev() int64 {
	return s.store.Rev()
}

// INFO: 把 WatchResponse send 到 ch, watchStream.Chan() 会监听这个 channel
func (w *watcher) send(watchResponse WatchResponse) bool {
	progressEvent := len(watchResponse.Events) == 0
	if progressEvent {
		return true
	}

	// INFO: Filter events
	if len(w.fcs) != 0 {
		events := make([]mvccpb.Event, 0, len(watchResponse.Events))
		for i := range watchResponse.Events {
			filtered := false
			for _, filter := range w.fcs {
				if filter(watchResponse.Events[i]) {
					filtered = true
					break
				}
			}
			if !filtered {
				events = append(events, watchResponse.Events[i])
			}
		}

		watchResponse.Events = events
	}

	// if all events are filtered out, we should send nothing.
	if !progressEvent && len(watchResponse.Events) == 0 {
		return true
	}

	select {
	case w.ch <- watchResponse:
		return true
	default:
		return false
	}
}
