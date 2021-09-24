package mvcc

import (
	"testing"

	"go.etcd.io/etcd/server/v3/lease"
	betesting "k8s-lx1036/k8s/storage/etcd/storage/backend/testing"
)

func TestStoreRev(t *testing.T) {
	b, _ := betesting.NewDefaultTmpBackend(t)
	s := NewStore(b, &lease.FakeLessor{}, StoreConfig{})
	defer s.Close()

	for i := 1; i <= 3; i++ {
		s.Put([]byte("foo"), []byte("bar"), lease.NoLease)
		if r := s.Rev(); r != int64(i+1) {
			t.Errorf("#%d: rev = %d, want %d", i, r, i+1)
		}
	}
}
