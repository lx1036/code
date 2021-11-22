package btree

import (
	"sort"
	"sync"
)

/*
INFO: btree 动态图 https://www.cs.usfca.edu/~galles/visualization/BTree.html
  代码写出一套 btree 还是挺难得！！！
*/

// Item INFO: btree 中一个节点存储 [m/2, m-1] 个 Item，且是 ordering 排序的, m 是 degree
type Item interface {
	Less(than Item) bool
}

// Int INFO: 整数型 Item
type Int int

func (a Int) Less(b Item) bool {
	return a < b.(Int)
}

type items []Item

func (s items) find(item Item) (index int, found bool) {
	i := sort.Search(len(s), func(i int) bool {
		return item.Less(s[i])
	})

	if i > 0 && !s[i-1].Less(item) {
		return i - 1, true
	}

	return i, false
}

type children []*node

type node struct {
	items items

	children children

	cow *copyOnWriteContext
}

func (n *node) get(key Item) Item {
	i, found := n.items.find(key)
	if found {
		return n.items[i]
	} else if len(n.children) > 0 { // 递归从 subtree 中查找，从 child node 中查找 item
		return n.children[i].get(key)
	}

	return nil
}

func (n *node) mutableFor(cow *copyOnWriteContext) *node {
	if n.cow == cow {
		return n
	}

}

// FreeList represents a free list of btree nodes. By default each
// BTree has its own FreeList, but multiple BTrees can share the same
// FreeList.
// Two Btrees using the same freelist are safe for concurrent write access.
// INFO: FreeList 专门分配 free node，而且还是线程安全的
type FreeList struct {
	mu       sync.Mutex
	freelist []*node
}

func NewFreeList(size int) *FreeList {
	return &FreeList{freelist: make([]*node, 0, size)}
}

func (f *FreeList) newNode() (n *node) {
	f.mu.Lock()
	index := len(f.freelist) - 1
	if index < 0 {
		f.mu.Unlock()
		return new(node)
	}

	n = f.freelist[index]
	f.freelist[index] = nil
	f.freelist = f.freelist[:index]
	f.mu.Unlock()

	return
}

type copyOnWriteContext struct {
	freelist *FreeList
}

func (c *copyOnWriteContext) newNode() (n *node) {
	n = c.freelist.newNode()
	n.cow = c
	return
}

const (
	DefaultFreeListSize = 32
)

type BTree struct {
	degree int
	length int

	root *node

	cow *copyOnWriteContext
}

func New(degree int) *BTree {
	return NewWithFreeList(degree, NewFreeList(DefaultFreeListSize))
}

func NewWithFreeList(degree int, f *FreeList) *BTree {
	if degree <= 1 {
		panic("bad degree")
	}

	return &BTree{
		degree: degree,
		cow:    &copyOnWriteContext{freelist: f},
	}
}

func (tree *BTree) ReplaceOrInsert(item Item) Item {
	if item == nil {
		panic("nil item being added to BTree")
	}

	if tree.root == nil {
		tree.root = tree.cow.newNode()
		tree.root.items = append(tree.root.items, item)
		tree.length++
		return nil
	} else {
		tree.root = tree.root.mutableFor(tree.cow)
		if len(tree.root.items) >= tree.maxItems() { // node split 分裂
			item2, second := tree.root.split(tree.maxItems() / 2)
			oldroot := tree.root
			tree.root = tree.cow.newNode()
			tree.root.items = append(tree.root.items, item2) // 中间 Item 为新 root
			tree.root.children = append(tree.root.children, oldroot, second)
		}
	}

	out := tree.root.insert(item, tree.maxItems())
	if out == nil {
		tree.length++
	}

	return out
}

func (tree *BTree) Get(key Item) Item {
	if tree.root == nil {
		return nil
	}

	return tree.root.get(key)
}

// maxItems returns the max number of items to allow per node.
func (tree *BTree) maxItems() int {
	return tree.degree*2 - 1
}
