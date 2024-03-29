package util

import "sync/atomic"

type AtomicUInt64 struct {
	v uint64
}

func (a *AtomicUInt64) Set(v uint64) {
	atomic.StoreUint64(&a.v, v)
}

func (a *AtomicUInt64) Get() uint64 {
	return atomic.LoadUint64(&a.v)
}
