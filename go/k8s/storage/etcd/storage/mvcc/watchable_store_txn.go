package mvcc

import (
	"go.etcd.io/etcd/api/v3/mvccpb"
)

// INFO: watchable 只有写事务

type WatchableKV interface {
	KV

	Watchable
}

type KV interface {
	ReadView
	WriteView

	// INFO: 创建写事务
	Write() TxnWrite

	// INFO: 创建读事务
	Read(mode ReadTxMode) TxnRead
}

type Watchable interface {
	// NewWatchStream returns a WatchStream that can be used to
	// watch events happened or happening on the KV.
	NewWatchStream() WatchStream
}

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
