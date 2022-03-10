package lru

// 双链表简介：https://leetcode-cn.com/leetbook/read/linked-list/jgk2s/

// 707: https://leetcode-cn.com/problems/design-linked-list/

type Node struct {
	prev, next *Node

	value interface{}
}

type List struct {
	len int

	head, tail *Node
}

// New returns an initialized list.
func NewList() *List {
	return new(List).Init()
}

func (l *List) Init() *List {
	l.len = 0

	l.head = &Node{}
	l.tail = &Node{}

	l.head.next = l.tail
	l.tail.prev = l.head

	return l
}

// get the value of index-th node in the linked-list
func (l *List) Get(index int) interface{} {
	if index < 0 || index >= l.len {
		return -1
	}

	var cur *Node
	if index < l.len/2 { // 从head开始查询
		cur = l.head
		// 空节点head的index是0
		for i := 0; i < index; i++ {
			cur = cur.next
		}
	} else { // 从tail开始查询
		cur = l.tail
		for i := l.len + 1; i > index; i-- {
			cur = cur.prev
		}
	}

	return cur.value
}

func (l *List) AddAtHead(value interface{}) {
	preHead, oldHead := l.head, l.head.next
	toAdd := &Node{
		prev:  preHead,
		next:  oldHead,
		value: value,
	}

	preHead.next = toAdd
	oldHead.prev = toAdd
	l.len++
}

func (l *List) AddAtTail(value interface{}) {
	preTail, oldTail := l.tail, l.tail.prev
	toAdd := &Node{
		prev:  oldTail,
		next:  preTail,
		value: value,
	}

	preTail.prev = toAdd
	oldTail.next = toAdd
	l.len++
}

func (l *List) AddAtIndex(index int, value interface{}) {
	if index > l.len {
		return
	}
	if index < 0 {
		index = 0
	}

	if index < l.len/2 { // 从头查询

	} else { //从尾查询

	}

	//toAdd := &Node{
	//	prev:  nil,
	//	next:  nil,
	//	value: value,
	//}

	l.len++
}

func (l *List) DeleteAtIndex(index int) {

}
