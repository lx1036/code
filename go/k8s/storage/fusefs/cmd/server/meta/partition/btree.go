package partition

import (
	"github.com/google/btree"
	"sync"
)

const DefaultBTreeDegree = 32

type BTree struct {
	sync.RWMutex
	tree *btree.BTree
}

// NewBtree creates a new btree.
func NewBtree() *BTree {
	return &BTree{
		tree: btree.New(DefaultBTreeDegree),
	}
}

func (b *BTree) GetTree() *BTree {
	b.Lock()
	defer b.Unlock()
	t := b.tree.Clone()
	nb := NewBtree()
	nb.tree = t
	return nb
}
