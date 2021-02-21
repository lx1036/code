package mvcc

import (
	"sync"

	"github.com/google/btree"
)

type treeIndex struct {
	sync.RWMutex

	tree *btree.BTree // b-tree 作为存储索引的数据结构
}

func (treeIdx *treeIndex) Put(key []byte, rev revision) {
	keyIdx := &keyIndex{key: key}

	treeIdx.Lock()
	defer treeIdx.Unlock()

	item := treeIdx.tree.Get(keyIdx)
	if item == nil {
		//keyIdx.Put(rev.main, rev.sub)
		treeIdx.tree.ReplaceOrInsert(keyIdx)
	}

	item.(*keyIndex).put(rev.main, rev.sub)
}

// 检查内存b-tree里是否存在keyIndex
func (treeIdx *treeIndex) keyIndex(keyIdx *keyIndex) *keyIndex {
	if item := treeIdx.tree.Get(keyIdx); item != nil {
		return item.(*keyIndex)
	}

	return nil
}

func (treeIdx *treeIndex) Get(key []byte, rev int64) (modified, created revision, ver int64, err error) {
	keyIdx := &keyIndex{key: key}

	treeIdx.Lock()
	defer treeIdx.Unlock()
	// 检查内存里b-tree是否存在keyIndex
	if keyIdx = treeIdx.keyIndex(keyIdx); keyIdx == nil {
		return revision{}, revision{}, 0, ErrRevisionNotFound
	}

	return keyIdx.get(rev)
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

func newTreeIndex() *treeIndex {
	return &treeIndex{
		tree: btree.New(32),
	}
}
