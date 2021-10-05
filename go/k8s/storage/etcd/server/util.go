package server

import (
	"encoding/binary"

	"k8s-lx1036/k8s/storage/etcd/raft"
)

type notifier struct {
	c   chan struct{}
	err error
}

func newNotifier() *notifier {
	return &notifier{
		c: make(chan struct{}),
	}
}

func (nc *notifier) notify(err error) {
	nc.err = err
	close(nc.c)
}

func uint64ToBigEndianBytes(number uint64) []byte {
	byteResult := make([]byte, 8)
	binary.BigEndian.PutUint64(byteResult, number)
	return byteResult
}

func isStopped(err error) bool {
	return err == raft.ErrStopped || err == ErrStopped
}
