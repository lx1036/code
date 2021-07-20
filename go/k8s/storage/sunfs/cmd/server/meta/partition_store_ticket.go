package meta

import (
	"github.com/google/btree"
)

type storeMsg struct {
	command    uint32
	applyIndex uint64
	inodeTree  *btree.BTree
	dentryTree *btree.BTree
}
