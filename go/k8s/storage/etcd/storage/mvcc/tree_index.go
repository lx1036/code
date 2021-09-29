package mvcc

import (
	"sort"
	"sync"

	"github.com/google/btree"
)

type TreeIndex interface {
	// INFO: 想要查询一个key的value，就必须指定版本号 revision
	Get(key []byte, atRev int64) (rev, created revision, ver int64, err error)
	Put(key []byte, rev revision)
	Range(key, end []byte, atRev int64) ([][]byte, []revision)
	RangeSince(key, end []byte, atRev int64) []revision
	Tombstone(key []byte, rev revision) error

	Revisions(key, end []byte, atRev int64, limit int) ([]revision, int)
	CountRevisions(key, end []byte, atRev int64) int
}

// TODO: Compact
type treeIndex struct {
	sync.RWMutex

	tree *btree.BTree // b-tree 作为存储索引的数据结构
}

func newTreeIndex() *treeIndex {
	return &treeIndex{
		tree: btree.New(32),
	}
}

func (treeIdx *treeIndex) Put(key []byte, rev revision) {
	treeIdx.Lock()
	defer treeIdx.Unlock()

	keyIdx := &keyIndex{key: key}
	item := treeIdx.tree.Get(keyIdx)
	if item == nil {
		keyIdx.put(rev.main, rev.sub)
		treeIdx.tree.ReplaceOrInsert(keyIdx)
		return
	}

	// item是指针，直接修改该item(keyIndex)值
	item.(*keyIndex).put(rev.main, rev.sub)
}

// 检查内存b-tree里是否存在keyIndex
func (treeIdx *treeIndex) keyIndex(keyIdx *keyIndex) *keyIndex {
	if item := treeIdx.tree.Get(keyIdx); item != nil {
		return item.(*keyIndex)
	}

	return nil
}

// INFO: 根据(key, revision.main)获取最新 revision
func (treeIdx *treeIndex) Get(key []byte, rev int64) (modified, created revision, ver int64, err error) {
	keyIdx := &keyIndex{key: key}

	treeIdx.Lock()
	defer treeIdx.Unlock()
	// 检查内存里b-tree是否存在keyIndex，根据key查找revision
	if keyIdx = treeIdx.keyIndex(keyIdx); keyIdx == nil {
		return revision{}, revision{}, 0, ErrRevisionNotFound
	}

	return keyIdx.get(rev)
}

// INFO: range 查询，没有锁
func (treeIdx *treeIndex) Range(key, end []byte, atRev int64) (keys [][]byte, revs []revision) {
	if end == nil {
		rev, _, _, err := treeIdx.Get(key, atRev)
		if err != nil {
			return nil, nil
		}
		return [][]byte{key}, []revision{rev}
	}

	treeIdx.visit(key, end, func(ki *keyIndex) bool {
		if rev, _, _, err := ki.get(atRev); err == nil {
			revs = append(revs, rev)
			keys = append(keys, ki.key)
		}
		return true
	})

	return keys, revs
}

// INFO: 利用了b+tree的range search范围查询
func (treeIdx *treeIndex) visit(key, end []byte, f func(ki *keyIndex) bool) {
	keyi, endi := &keyIndex{key: key}, &keyIndex{key: end}

	treeIdx.RLock()
	defer treeIdx.RUnlock()

	treeIdx.tree.AscendGreaterOrEqual(keyi, func(item btree.Item) bool {
		if len(endi.key) > 0 && !item.Less(endi) {
			return false
		}
		if !f(item.(*keyIndex)) {
			return false
		}

		return true
	})
}

// INFO: range 查询，有锁
func (treeIdx *treeIndex) RangeSince(key, end []byte, atRev int64) (revs []revision) {
	keyi := &keyIndex{key: key}

	treeIdx.RLock()
	defer treeIdx.RUnlock()

	if end == nil {
		item := treeIdx.tree.Get(keyi)
		if item == nil {
			return nil
		}
		keyi = item.(*keyIndex)
		return keyi.since(atRev)
	}

	endi := &keyIndex{key: end}
	treeIdx.tree.AscendGreaterOrEqual(keyi, func(item btree.Item) bool {
		if len(endi.key) > 0 && !item.Less(endi) {
			return false
		}
		curKeyi := item.(*keyIndex)
		revs = append(revs, curKeyi.since(atRev)...)
		return true
	})

	sort.Sort(revisions(revs))

	return revs
}

func (treeIdx *treeIndex) Tombstone(key []byte, rev revision) error {
	keyIdx := &keyIndex{key: key}

	treeIdx.Lock()
	defer treeIdx.Unlock()
	item := treeIdx.tree.Get(keyIdx)
	if item == nil {
		return ErrRevisionNotFound
	}

	return item.(*keyIndex).tombstone(rev.main, rev.sub)
}

// Revisions INFO: [key, end] 之间的 revisions，且 revision.main > atRev, 且 len(revisions)<=limit
func (treeIdx *treeIndex) Revisions(key, end []byte, atRev int64, limit int) (revs []revision, total int) {
	if end == nil {
		rev, _, _, err := treeIdx.Get(key, atRev)
		if err != nil {
			return nil, 0
		}

		return []revision{rev}, 1
	}

	treeIdx.visit(key, end, func(ki *keyIndex) bool {
		if rev, _, _, err := ki.get(atRev); err == nil {
			if limit <= 0 || len(revs) < limit {
				revs = append(revs, rev)
			}
			total++
		}
		return true
	})

	return revs, total
}

// CountRevisions INFO: 这里返回 len(revisions)
func (treeIdx *treeIndex) CountRevisions(key, end []byte, atRev int64) int {
	if end == nil {
		_, _, _, err := treeIdx.Get(key, atRev)
		if err != nil {
			return 0
		}
		return 1
	}

	total := 0
	treeIdx.visit(key, end, func(ki *keyIndex) bool {
		if _, _, _, err := ki.get(atRev); err == nil {
			total++
		}
		return true
	})

	return total
}
