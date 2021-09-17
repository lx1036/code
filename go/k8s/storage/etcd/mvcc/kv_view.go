package mvcc

import (
	"go.etcd.io/etcd/server/v3/lease"
)

type readView struct{ kv KV }

type writeView struct{ kv KV }

func (wv *writeView) Put(key, value []byte, lease lease.LeaseID) (rev int64) {
	tw := wv.kv.Write()
	defer tw.End()

	return tw.Put(key, value, lease)
}
