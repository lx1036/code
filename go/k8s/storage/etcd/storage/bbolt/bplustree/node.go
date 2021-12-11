package bplustree

import (
	"fmt"
	"k8s.io/klog/v2"
	"sort"
)

const (
	MaxKV = 255
	//MaxKV = 5
	MaxKC = 511
	//MaxKC = 5
)

type node interface {
	find(key int) (int, bool)
	parent() *branchNode
	setParent(*branchNode)
	full() bool
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
