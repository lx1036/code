package mvcc

import (
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/server/v3/lease"
)

type ReadTxMode uint32

const (
	// Use ConcurrentReadTx and the txReadBuffer is copied
	ConcurrentReadTxMode = ReadTxMode(1)
	// Use backend ReadTx and txReadBuffer is not copied
	SharedBufReadTxMode = ReadTxMode(2)
)

type WatchableKV interface {
	KV

	Watchable
}

type KV interface {

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

type ReadView interface {

	// Rev returns the revision of the KV at the time of opening the txn.
	Rev() int64
}

type WriteView interface {
	Put(key, value []byte, lease lease.LeaseID) (rev int64)

	// DeleteRange
	// INFO: 范围删除，会触发 delete event
	//  delete keys in [key, end)
	//  如果end is nil, delete the key
	DeleteRange(key, end []byte) (n, rev int64)
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

	// Changes gets the changes made since opening the write txn.
	Changes() []mvccpb.KeyValue
}
