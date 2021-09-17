package mvcc

import (
	"go.etcd.io/etcd/server/v3/lease"
)

type WatchableKV interface {
	KV

	Watchable
}

type KV interface {

	// Write creates a write transaction.
	Write() TxnWrite
}

type Watchable interface {
	// NewWatchStream returns a WatchStream that can be used to
	// watch events happened or happening on the KV.
	NewWatchStream() WatchStream
}

type ReadView interface {

	// Rev returns the revision of the KV at the time of opening the txn.
	Rev() int64
}

type WriteView interface {
	Put(key, value []byte, lease lease.LeaseID) (rev int64)
}

// INFO: 只读事务
type TxnRead interface {
	ReadView

	// End marks the transaction is complete and ready to commit.
	End()
}

// INFO: 读写事务
type TxnWrite interface {
	TxnRead

	WriteView
}
