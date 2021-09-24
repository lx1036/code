package backend

import (
	"testing"

	betesting "k8s-lx1036/k8s/storage/etcd/storage/backend/testing"
)

var (
	testBucketName = []byte("test")
)

var (
	Test = bucket{id: 100, name: testBucketName, safeRangeBucket: false}
)

type bucket struct {
	id              BucketID
	name            []byte
	safeRangeBucket bool
}

func (b bucket) ID() BucketID {
	return b.id
}

func (b bucket) Name() []byte {
	return b.name
}

func (b bucket) String() string {
	return string(b.Name())
}

func (b bucket) IsSafeRangeBucket() bool {
	return b.safeRangeBucket
}

func TestSnapshot(t *testing.T) {
	b, _ := betesting.NewDefaultTmpBackend(t)
	defer b.Close()

	tx := b.BatchTx()
	tx.Lock()
	tx.UnsafeCreateBucket(Test)
	tx.UnsafePut(Test, []byte("foo"), []byte("bar"))
	tx.Unlock()
	b.ForceCommit()

}
