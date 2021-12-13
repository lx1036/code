package bplustree

import (
	"sort"
)

type BTree struct {
	root   *branchNode
	first  *leafNode
	leaf   int
	branch int
	height int
}

// 初始化一个 degree=2 的 b+tree
func New() *BTree {
	leaf := newLeafNode(nil)
	branch := newBranchNode(nil, leaf)
	leaf.p = branch

	return &BTree{
		root:   branch,
		first:  leaf,
		leaf:   1,
		branch: 1,
		height: 2,
	}
}

// Insert inserts a (key, value) into the B+ tree
func (tree *BTree) Insert(key int, value string) {
	_, oldIndex, leaf := search(tree.root, key)
	branch := leaf.parent()

	mid, split := leaf.insert(key, value)
	if !split {
		return
	}

	var midNode node
	midNode = leaf
	branch.inodes[oldIndex].child = leaf.next
	leaf.next.setParent(branch)
	interior, interiorP := branch, branch.parent()

	for {
		oldIndex := 0
		newNode := new(branchNode)

		isRoot := interiorP == nil
		if !isRoot {
			oldIndex, _ = interiorP.find(key)
		}

		mid, newNode, split = interior.insert(mid, midNode)
		if !split {
			return
		}

		if !isRoot {
			interiorP.inodes[oldIndex].child = newNode
			newNode.setParent(interiorP)

			midNode = interior
		} else {
			tree.root = newBranchNode(nil, newNode)
			newNode.setParent(tree.root)

			tree.root.insert(mid, interior)
			return
		}

		interior, interiorP = interiorP, interiorP.parent()
	}
}

// Search searches the key in B+ tree
// If the key exists, it returns the value of key and true
// If the key does not exist, it returns an empty string and false
func (tree *BTree) Search(key int) (string, bool) {
	kv, _, _ := search(tree.root, key)
	if kv == nil {
		return "", false
	}
	return kv.value, true
}

func search(n node, key int) (*kv, int, *leafNode) {
	curr := n
	oldIndex := -1

	for {
		switch t := curr.(type) {
		case *leafNode:
			i, ok := t.find(key)
			if !ok {
				return nil, oldIndex, t
			}
			return &t.kvs[i], oldIndex, t
		case *branchNode:
			i, _ := t.find(key)
			curr = t.inodes[i].child // INFO: 这里为何是 "比 key 大的最小 inodes[i].key"
			oldIndex = i
		default:
			panic("")
		}
	}
}

const (
	bucketLeafFlag   = 0x01
	bucketBranchFlag = 0x02
)

// branch node 只存储 key
type kc struct {
	key   int
	child node
}

type inode struct {
	flags int
	key   int
	value string
	child node // INFO: child 在每一个 kv 里设置了
}

// one empty slot for split
type inodes [MaxKC + 1]inode

func (a *inodes) Len() int           { return len(a) }
func (a *inodes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a *inodes) Less(i, j int) bool { return a[i].key < a[j].key }

type branchNode struct {
	inodes

	count int
	p     *branchNode
}

func newBranchNode(p *branchNode, largestChild node) *branchNode {
	b := &branchNode{
		count: 1,
		p:     p,
	}

	if largestChild != nil {
		b.inodes[0].child = largestChild
	}

	return b
}

// INFO: 比 key 大的最小 inodes[i].key
func (b *branchNode) find(key int) (int, bool) {
	i := sort.Search(b.count-1, func(i int) bool {
		return b.inodes[i].key > key
	})

	return i, true
}

// INFO: key 插入该 branch node，还得指定 child node
func (b *branchNode) insert(key int, child node) (int, *branchNode, bool) {
	i, _ := b.find(key)

	// 节点数量还未到 m-1，m 是 degree，在 i 处后移一位
	if !b.full() {
		copy(b.inodes[i+1:], b.inodes[i:b.count])
		b.inodes[i].key = key
		b.inodes[i].child = child
		child.setParent(b)
		b.count++
		return 0, nil, false
	}

	// insert the new node into the empty slot
	b.inodes[MaxKC].key = key
	b.inodes[MaxKC].child = child
	child.setParent(b)
	next, midKey := b.split()

	return midKey, next, true
}

func (b *branchNode) split() (*branchNode, int) {
	sort.Sort(&b.inodes)

	// get the mid info
	midIndex := MaxKC / 2
	midChild := b.inodes[midIndex].child
	midKey := b.inodes[midIndex].key

	// create the split node with out a parent
	next := newBranchNode(nil, nil)
	copy(next.inodes[0:], b.inodes[midIndex+1:]) // 后半部分
	next.count = MaxKC - midIndex
	// update parent
	for i := 0; i < next.count; i++ {
		next.inodes[i].child.setParent(next) // 更新新的 branchNode 的 child.parent
	}

	// modify the original node
	b.count = midIndex + 1
	//b.kcs[b.count-1].key = 0
	//b.kcs[b.count-1].child = midChild
	midChild.setParent(b)

	return next, midKey
}

func (b *branchNode) parent() *branchNode {
	return b.p
}

func (b *branchNode) setParent(p *branchNode) {
	b.p = p
}

func (b *branchNode) full() bool {
	return b.count == MaxKC
}
