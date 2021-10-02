package backend

import (
	"fmt"
	"io"

	bolt "go.etcd.io/bbolt"
	"k8s.io/klog/v2"
)

type Snapshot interface {
	// Size gets the size of the snapshot.
	Size() int64
	// WriteTo writes the snapshot into the given writer.
	WriteTo(w io.Writer) (n int64, err error)
	// Close closes the snapshot.
	Close() error
}

type snapshot struct {
	*bolt.Tx
	stopc chan struct{}
	donec chan struct{}
}

func (b *backend) Snapshot() Snapshot {
	// INFO: 为何先 commit? 保证数据已经落盘
	b.batchTx.Commit()

	// read-only lock
	b.RLock()
	defer b.RUnlock()
	tx, err := b.db.Begin(false) // read-only
	if err != nil {
		klog.Fatalf(fmt.Sprintf("[Snapshot]begin transaction err %v", err))
	}
	stopc, donec := make(chan struct{}), make(chan struct{})
	dbBytes := tx.Size() // returns current database size in bytes as seen by this transaction
	//kb := 1024 * 1024
	klog.Infof(fmt.Sprintf("[Snapshot]db size %d bytes", int64(dbBytes)))

	return &snapshot{
		Tx:    tx,
		stopc: stopc,
		donec: donec,
	}
}

// Close INFO: Close 里去 Rollback
func (s *snapshot) Close() error {
	close(s.stopc)
	<-s.donec

	return s.Tx.Rollback()
}
