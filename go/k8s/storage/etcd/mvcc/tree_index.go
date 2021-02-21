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

func (treeIdx *treeIndex) Tombstone(key []byte, rev revision) error {
	keyIdx := &keyIndex{key: key}

	treeIdx.Lock()
	defer treeIdx.Unlock()
	item := treeIdx.tree.Get(keyIdx)
	if item == nil {
		return ErrRevisionNotFound
	}

	item.(*keyIndex).tombstone(rev.main, rev.sub)
}

func newTreeIndex() *treeIndex {
	return &treeIndex{
		tree: btree.New(32),
	}
}
