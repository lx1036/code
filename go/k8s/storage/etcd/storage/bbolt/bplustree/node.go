package bplustree

import (
	"fmt"
	"k8s.io/klog/v2"
	"sort"
)

const (
	MaxKV = 255
	MaxKC = 511
)

type node interface {
	find(key int) (int, bool)
	parent() *branchNode
	setParent(*branchNode)
	full() bool
}

// branch node 只存储 key
type kc struct {
	key   int
	child node
}

// one empty slot for split
type kcs [MaxKC + 1]kc

func (a *kcs) Len() int      { return len(a) }
func (a *kcs) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a *kcs) Less(i, j int) bool {
	if a[i].key == 0 {
		return false
	}

	if a[j].key == 0 {
		return true
	}

	return a[i].key < a[j].key
}

type branchNode struct {
	kcs kcs

	count int
	p     *branchNode
}

func newBranchNode(p *branchNode, largestChild node) *branchNode {
	b := &branchNode{
		count: 1,
		p:     p,
	}

	if largestChild != nil {
		b.kcs[0].child = largestChild
	}

	return b
}

func (b *branchNode) find(key int) (int, bool) {
	i := sort.Search(b.count-1, func(i int) bool {
		return b.kcs[i].key > key
	})

	return i, true
}

func (b *branchNode) insert(key int, child node) (int, *branchNode, bool) {
	i, _ := b.find(key)

	// 节点数量还未到 m-1，m 是 degree，在 i 处后移一位
	if !b.full() {
		copy(b.kcs[i+1:], b.kcs[i:b.count])
		b.kcs[i].key = key
		b.kcs[i].child = child
		child.setParent(b)
		b.count++
		return 0, nil, false
	}

	// insert the new node into the empty slot
	b.kcs[MaxKC].key = key
	b.kcs[MaxKC].child = child
	child.setParent(b)
	next, midKey := b.split()

	return midKey, next, true
}

func (b *branchNode) split() (*branchNode, int) {
	sort.Sort(&b.kcs)

	// get the mid info
	midIndex := MaxKC / 2
	midChild := b.kcs[midIndex].child
	midKey := b.kcs[midIndex].key

	// create the split node with out a parent
	next := newBranchNode(nil, nil)
	copy(next.kcs[0:], b.kcs[midIndex+1:]) // 后半部分
	next.count = MaxKC - midIndex
	// update parent
	for i := 0; i < next.count; i++ {
		next.kcs[i].child.setParent(next)
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

type kv struct {
	key   int
	value string
}

type kvs [MaxKV]kv

func (a *kvs) Len() int           { return len(a) }
func (a *kvs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a *kvs) Less(i, j int) bool { return a[i].key < a[j].key }

type leafNode struct {
	kvs   kvs
	count int
	next  *leafNode

	p *branchNode // 指向 branch node 指针
}

func newLeafNode(p *branchNode) *leafNode {
	return &leafNode{
		p: p,
	}
}

// find finds the index of a key in the leaf node.
// If the key exists in the node, it returns the index and true.
// If the key does not exist in the node, it returns index to
// insert the key (the index of the smallest key in the node that larger
// than the given key) and false.
func (l *leafNode) find(key int) (int, bool) {
	i := sort.Search(l.count, func(i int) bool {
		return l.kvs[i].key >= key
	})

	if i < l.count && l.kvs[i].key == key {
		return i, true
	}

	return i, false
}

// insert 如果分裂，则返回middle key
func (l *leafNode) insert(key int, value string) (int, bool) {
	i, ok := l.find(key)

	if ok {
		klog.Infof(fmt.Sprintf("existed key %d at index %d", key, i))
		l.kvs[i].value = value
		return 0, false
	}

	// 节点数量还未到 m-1，m 是 degree，在 i 处后移一位
	if !l.full() {
		copy(l.kvs[i+1:], l.kvs[i:l.count])
		l.kvs[i].key = key
		l.kvs[i].value = value
		l.count++
		return 0, false
	}

	// 节点已经满了，分裂
	next := l.split()

	if key < next.kvs[0].key {
		l.insert(key, value)
	} else {
		next.insert(key, value)
	}

	return next.kvs[0].key, true
}

// 后一半移走
func (l *leafNode) split() *leafNode {
	next := newLeafNode(nil)
	copy(next.kvs[0:], l.kvs[l.count/2+1:])
	next.count = MaxKV - l.count/2 - 1
	next.next = l.next // ???
	l.count = l.count/2 + 1
	l.next = next

	return next
}

func (l *leafNode) parent() *branchNode {
	return l.p
}

func (l *leafNode) setParent(p *branchNode) {
	l.p = p
}

func (l *leafNode) full() bool {
	return l.count == MaxKV
}
