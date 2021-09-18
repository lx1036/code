package mvcc

import (
	"bytes"
	"errors"
	"sync"

	"go.etcd.io/etcd/api/v3/mvccpb"
)

const AutoWatchID WatchID = 0

var (
	ErrWatcherNotExist    = errors.New("mvcc: watcher does not exist")
	ErrEmptyWatcherRange  = errors.New("mvcc: watcher range is empty")
	ErrWatcherDuplicateID = errors.New("mvcc: duplicate watch ID provided on the WatchStream")
)

type WatchID int64

// FilterFunc returns true if the given event should be filtered out.
type FilterFunc func(e mvccpb.Event) bool

type WatchResponse struct {
	// WatchID is the WatchID of the watcher this response sent to.
	WatchID WatchID

	// Events contains all the events that needs to send.
	Events []mvccpb.Event

	// Revision is the revision of the KV when the watchResponse is created.
	// For a normal response, the revision should be the same as the last
	// modified revision inside Events. For a delayed response to a unsynced
	// watcher, the revision is greater than the last modified revision
	// inside Events.
	Revision int64

	// CompactRevision is set when the watcher is cancelled due to compaction.
	CompactRevision int64
}

type WatchStream interface {

	// Watch INFO: 创建一个watcher, watch events keys [key, end)
	Watch(id WatchID, key, end []byte, startRev int64, fcs ...FilterFunc) (WatchID, error)

	// Chan returns a chan. All watch response will be sent to the returned chan.
	Chan() <-chan WatchResponse

	// Cancel cancels a watcher by giving its ID. If watcher does not exist, an error will be
	// returned.
	Cancel(id WatchID) error

	// Close closes Chan and release all related resources.
	Close()
}

// INFO:
//  watchStream contains a collection of watchers that share
//  one streaming chan to send out watched events and other control events.
type watchStream struct {
	mu sync.Mutex

	watchable watchable

	ch chan WatchResponse

	cancels  map[WatchID]cancelFunc
	watchers map[WatchID]*watcher

	// nextID is the ID pre-allocated for next new watcher in this stream
	nextID WatchID

	closed bool
}

func (s *watchableStore) NewWatchStream() WatchStream {
	return &watchStream{
		watchable: s,
		ch:        make(chan WatchResponse, chanBufLen),
		cancels:   make(map[WatchID]cancelFunc),

		watchers: make(map[WatchID]*watcher),
	}
}

// Watch INFO: 创建stream中的一个watcher，并返回WatchID
func (ws *watchStream) Watch(id WatchID, key, end []byte, startRev int64, fcs ...FilterFunc) (WatchID, error) {
	// 必须 key < end
	if len(end) != 0 && bytes.Compare(key, end) != -1 {
		return -1, ErrEmptyWatcherRange
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.closed {
		return -1, ErrEmptyWatcherRange
	}

	if id == AutoWatchID {
		for ws.watchers[ws.nextID] != nil {
			ws.nextID++
		}
		id = ws.nextID
		ws.nextID++
	} else if _, ok := ws.watchers[id]; ok {
		return -1, ErrWatcherDuplicateID
	}

	watcher, cancelFunc := ws.watchable.watch(key, end, startRev, id, ws.ch, fcs...)
	ws.cancels[id] = cancelFunc
	ws.watchers[id] = watcher

	return id, nil
}

func (ws *watchStream) Chan() <-chan WatchResponse {
	return ws.ch
}

func (ws *watchStream) Cancel(id WatchID) error {
	ws.mu.Lock()
	cancel, ok := ws.cancels[id]
	w := ws.watchers[id]
	ok = ok && !ws.closed
	ws.mu.Unlock()

	if !ok {
		return ErrWatcherNotExist
	}
	cancel()

	// TODO: 这块为何需要判断???

	ws.mu.Lock()
	// The watch isn't removed until cancel so that if Close() is called,
	// it will wait for the cancel. Otherwise, Close() could close the
	// watch channel while the store is still posting events.
	if ww := ws.watchers[id]; ww == w {
		delete(ws.cancels, id)
		delete(ws.watchers, id)
	}
	ws.mu.Unlock()

	return nil
}

func (ws *watchStream) Close() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for _, cancel := range ws.cancels {
		cancel()
	}

	ws.closed = true
	close(ws.ch)
}
