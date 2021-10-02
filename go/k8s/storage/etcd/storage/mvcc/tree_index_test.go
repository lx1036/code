package mvcc

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	betesting "k8s-lx1036/k8s/storage/etcd/storage/backend/testing"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/server/v3/lease"
	"k8s.io/klog/v2"
)

func newTestKeyIndex() *keyIndex {
	// key: "foo"
	// rev: 16
	// generations:
	//    {empty}
	//    {{14, 0}[1], {14, 1}[2], {16, 0}(t)[3]}
	//    {{8, 0}[1], {10, 0}[2], {12, 0}(t)[3]}
	//    {{2, 0}[1], {4, 0}[2], {6, 0}(t)[3]}

	keyIdx := &keyIndex{key: []byte("foo")}
	keyIdx.put(2, 0)
	keyIdx.put(4, 0)
	keyIdx.tombstone(6, 0)
	keyIdx.put(8, 0)
	keyIdx.put(10, 0)
	keyIdx.tombstone(12, 0)
	keyIdx.put(14, 0)
	keyIdx.put(14, 1)
	keyIdx.tombstone(16, 0)

	klog.Infof("modified revision main:%d sub:%d", keyIdx.modified.main, keyIdx.modified.sub)

	return keyIdx
}

// 查看KeyIndex，即generation是否为空
func TestKeyIndexIsEmpty(test *testing.T) {
	fixtures := []struct {
		keyIdx *keyIndex
		w      bool
	}{
		{
			&keyIndex{
				key:         []byte("foo"),
				generations: []generation{{}},
			},
			true,
		},
		{
			&keyIndex{
				key:      []byte("foo"),
				modified: revision{2, 0},
				generations: []generation{
					{created: revision{1, 0}, version: 2, revisions: []revision{{main: 2}}},
				},
			},
			false,
		},
	}

	for i, fixture := range fixtures {
		g := fixture.keyIdx.isEmpty()
		if g != fixture.w {
			test.Errorf("#%d: isEmpty = %v, want %v", i, g, fixture.w)
		}
	}
}

// 测试findGeneration()
func TestKeyIndexFindGeneration(test *testing.T) {
	keyIdx := newTestKeyIndex()

	fixtures := []struct {
		rev int64
		wg  *generation
	}{
		{0, nil},
		{1, nil},
		{2, &keyIdx.generations[0]},
		{3, &keyIdx.generations[0]},
		{4, &keyIdx.generations[0]},
		{5, &keyIdx.generations[0]},
		{6, nil},
		{7, nil},
		{8, &keyIdx.generations[1]},
		{9, &keyIdx.generations[1]},
		{10, &keyIdx.generations[1]},
		{11, &keyIdx.generations[1]},
		{12, nil},
		{13, nil},
	}

	for i, fixture := range fixtures {
		g := keyIdx.findGeneration(fixture.rev)
		if g != fixture.wg {
			test.Errorf("#%d: generation = %+v, want %+v", i, g, fixture.wg)
		}
	}
}

func TestTreeIndexGet(test *testing.T) {
	treeIdx := newTreeIndex()
	treeIdx.Put([]byte("foo"), revision{main: 2})
	treeIdx.Put([]byte("foo"), revision{main: 4})
	treeIdx.Tombstone([]byte("foo"), revision{main: 6})

	fixtures := []struct {
		desc      string
		rev       int64
		wantedRev revision
		created   revision
		wver      int64
		werr      error
	}{
		{"revision=0", 0, revision{}, revision{}, 0, ErrRevisionNotFound},
		{"revision=1", 1, revision{}, revision{}, 0, ErrRevisionNotFound},
		{"revision=2", 2, revision{main: 2}, revision{main: 2}, 1, nil},
		{"revision=3", 3, revision{main: 2}, revision{main: 2}, 1, nil},
		{"revision=4", 4, revision{main: 4}, revision{main: 2}, 2, nil},
		{"revision=5", 5, revision{main: 4}, revision{main: 2}, 2, nil},
		{"revision=6", 6, revision{}, revision{}, 0, ErrRevisionNotFound},
	}

	for i, fixture := range fixtures {
		test.Run(fixture.desc, func(t *testing.T) {
			rev, created, ver, err := treeIdx.Get([]byte("foo"), fixture.rev)
			if err != fixture.werr {
				test.Errorf("#%d: err = %v, want %v", i, err, fixture.werr)
			}
			if rev != fixture.wantedRev {
				test.Errorf("#%d: rev = %+v, want %+v", i, rev, fixture.wantedRev)
			}
			if created != fixture.created {
				test.Errorf("#%d: created = %+v, want %+v", i, created, fixture.created)
			}
			if ver != fixture.wver {
				test.Errorf("#%d: ver = %d, want %d", i, ver, fixture.wver)
			}
		})
	}
}

func TestTreeIndexRange(test *testing.T) {
	treeIdx := newTreeIndex()
	allKeys := [][]byte{[]byte("foo"), []byte("foo1"), []byte("foo2")}
	allRevs := []revision{{main: 1}, {main: 2}, {main: 3}}
	for i := range allKeys {
		treeIdx.Put(allKeys[i], allRevs[i])
	}

	fixtures := []struct {
		desc     string
		key, end []byte
		wkeys    [][]byte
		wrevs    []revision
	}{
		{
			"single key that not found", []byte("bar"), nil, nil, nil,
		},
		{
			"single key that found", []byte("foo"), nil, allKeys[:1], allRevs[:1],
		},
		{
			"range keys, return first member", []byte("foo"), []byte("foo1"), allKeys[:1], allRevs[:1],
		},
		{
			"range keys, return first two members", []byte("foo"), []byte("foo2"), allKeys[:2], allRevs[:2],
		},
		{
			"range keys, return all members", []byte("foo"), []byte("fop"), allKeys, allRevs,
		},
		{
			"range keys, return last two members", []byte("foo1"), []byte("fop"), allKeys[1:], allRevs[1:],
		},
		{
			"range keys, return last member", []byte("foo2"), []byte("fop"), allKeys[2:], allRevs[2:],
		},
		{
			"range keys, return nothing", []byte("foo3"), []byte("fop"), nil, nil,
		},
	}

	atRev := int64(3)
	for i, fixture := range fixtures {
		test.Run(fixture.desc, func(t *testing.T) {
			keys, revs := treeIdx.Range(fixture.key, fixture.end, atRev)
			if !reflect.DeepEqual(keys, fixture.wkeys) {
				t.Errorf("#%d: keys = %+v, want %+v", i, keys, fixture.wkeys)
			}
			if !reflect.DeepEqual(revs, fixture.wrevs) {
				t.Errorf("#%d: revs = %+v, want %+v", i, revs, fixture.wrevs)
			}
		})
	}
}

func TestTreeIndexRangeSince(test *testing.T) {
	treeIdx := newTreeIndex()
	allKeys := [][]byte{[]byte("foo"), []byte("foo1"), []byte("foo2"), []byte("foo2"), []byte("foo1"), []byte("foo")}
	allRevs := []revision{{main: 1}, {main: 2}, {main: 3}, {main: 4}, {main: 5}, {main: 6}}
	for i := range allKeys {
		treeIdx.Put(allKeys[i], allRevs[i])
	}

	fixtures := []struct {
		desc     string
		key, end []byte
		wrevs    []revision
	}{
		{
			"single key that not found", []byte("bar"), nil, nil,
		},
		{
			"single key that found", []byte("foo"), nil, []revision{{main: 1}, {main: 6}},
		},
		{
			"range keys, return first member", []byte("foo"), []byte("foo1"), []revision{{main: 1}, {main: 6}},
		},
		{
			"range keys, return first two members", []byte("foo"), []byte("foo2"), []revision{{main: 1}, {main: 2}, {main: 5}, {main: 6}},
		},
		{
			"range keys, return all members", []byte("foo"), []byte("fop"), allRevs,
		},
		{
			"range keys, return last two members", []byte("foo1"), []byte("fop"), []revision{{main: 2}, {main: 3}, {main: 4}, {main: 5}},
		},
		{
			"range keys, return last member", []byte("foo2"), []byte("fop"), []revision{{main: 3}, {main: 4}},
		},
		{
			"range keys, return nothing", []byte("foo3"), []byte("fop"), nil,
		},
	}

	atRev := int64(1)
	for i, fixture := range fixtures {
		test.Run(fixture.desc, func(t *testing.T) {
			revs := treeIdx.RangeSince(fixture.key, fixture.end, atRev)
			if !reflect.DeepEqual(revs, fixture.wrevs) {
				t.Errorf("#%d: revs = %+v, want %+v", i, revs, fixture.wrevs)
			}
		})
	}
}

func TestTreeIndexTombstone(test *testing.T) {
	treeIdx := newTreeIndex()
	treeIdx.Put([]byte("foo"), revision{main: 1})

	err := treeIdx.Tombstone([]byte("foo"), revision{main: 2})
	if err != nil {
		test.Errorf("tombstone error = %v, want nil", err)
	}

	_, _, _, err = treeIdx.Get([]byte("foo"), 2)
	if err != ErrRevisionNotFound {
		test.Errorf("get error = %v, want ErrRevisionNotFound", err)
	}

	err = treeIdx.Tombstone([]byte("foo"), revision{main: 3})
	if err != ErrRevisionNotFound {
		test.Errorf("tombstone error = %v, want %v", err, ErrRevisionNotFound)
	}
}

func TestStoreRev(t *testing.T) {
	// INFO: (1) basic read/write
	b, tmpPath := betesting.NewDefaultTmpBackend(t)
	s := NewStore(b, &lease.FakeLessor{}, StoreConfig{})
	defer s.Close()
	defer os.RemoveAll(tmpPath)

	for i := 1; i <= 3; i++ {
		s.Put([]byte("foo"), []byte("bar"), lease.NoLease)
		// store current revision: 2,3,4, store启动时默认初始是 1
		if r := s.Rev(); r != int64(i+1) {
			t.Errorf("#%d: rev = %d, want %d", i, r, i+1)
		}

		result, err := s.Range(context.TODO(), []byte("foo"), nil, RangeOptions{})
		if err != nil {
			klog.Fatal(err)
		}

		for _, keyValue := range result.KVs {
			klog.Infof(fmt.Sprintf("key: %s, value: %s, mod revision: %d", string(keyValue.Key), string(keyValue.Value), keyValue.ModRevision))
		}
	}

	deletedKeysLen, rev := s.DeleteRange([]byte("foo"), []byte("goo"))
	klog.Infof(fmt.Sprintf("deletedKeysLen: %d, revision: %d", deletedKeysLen, rev))
	result, err := s.Range(context.TODO(), []byte("foo"), nil, RangeOptions{})
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infof(fmt.Sprintf("%+v", *result))

	// INFO: (2)ensures Read does not blocking Write after its creation
	// write something to read later
	s.Put([]byte("foo1"), []byte("bar"), lease.NoLease)
	// readTx simulates a long read request
	concurrentReadTx1 := s.Read(ConcurrentReadTxMode)
	// write should not be blocked by reads
	done := make(chan struct{}, 1)
	go func() {
		s.Put([]byte("foo1"), []byte("newBar"), lease.NoLease) // this is a write Txn
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatalf("write should not be blocked by read")
	}
	// readTx2 simulates a short read request
	concurrentReadTx2 := s.Read(ConcurrentReadTxMode)
	ro := RangeOptions{Limit: 1, Rev: 0, Count: false}
	result, err = concurrentReadTx2.Range(context.TODO(), []byte("foo1"), nil, ro)
	if err != nil {
		t.Fatalf("failed to range: %v", err)
	}
	// readTx2 should see the result of new write
	w := mvccpb.KeyValue{
		Key:            []byte("foo1"),
		Value:          []byte("newBar"),
		CreateRevision: 6, // foo1 创建时已经是 6 版本
		ModRevision:    7, // INFO: 6 次写操作 + 1, 包含 goroutine 那一次
		Version:        2, // 已经写操作了 2 次
	}
	if !reflect.DeepEqual(result.KVs[0], w) {
		t.Fatalf("range result = %+v, want = %+v", result.KVs[0], w)
	}
	concurrentReadTx2.End()
	result, err = concurrentReadTx1.Range(context.TODO(), []byte("foo1"), nil, ro)
	if err != nil {
		t.Fatalf("failed to range: %v", err)
	}
	// readTx1 should not see the result of new write
	w = mvccpb.KeyValue{
		Key:            []byte("foo1"),
		Value:          []byte("bar"),
		CreateRevision: 6,
		ModRevision:    6, // INFO: concurrentReadTx1 时只有 5 次写操作，没有 goroutine 那一次
		Version:        1,
	}
	if !reflect.DeepEqual(result.KVs[0], w) {
		t.Fatalf("range result = %+v, want = %+v", result.KVs[0], w)
	}
	concurrentReadTx1.End()

	// INFO: (3)creates random concurrent Reads and Writes, and ensures Reads always see latest Writes

}
