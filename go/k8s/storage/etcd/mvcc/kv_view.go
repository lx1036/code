package mvcc

import (
	"go.etcd.io/etcd/server/v3/lease"
)

type readView struct{ kv KV }

func (rv *readView) Rev() int64 {
	tr := rv.kv.Read(ConcurrentReadTxMode)
	defer tr.End()
	return tr.Rev()
}

type writeView struct{ kv KV }

func (wv *writeView) Put(key, value []byte, lease lease.LeaseID) (rev int64) {
	tw := wv.kv.Write()
	defer tw.End()

	return tw.Put(key, value, lease)
}

func (wv *writeView) DeleteRange(key, end []byte) (n, rev int64) {
	tw := wv.kv.Write()
	defer tw.End()

	return tw.DeleteRange(key, end)
}
