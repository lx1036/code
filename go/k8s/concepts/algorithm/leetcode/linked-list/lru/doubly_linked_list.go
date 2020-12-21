package lru

// 双链表简介：https://leetcode-cn.com/leetbook/read/linked-list/jgk2s/

type Node struct {
	prev, next *Node

	value interface{}
}

type List struct {
	len int

	head, tail *Node
}

// New returns an initialized list.
func New() *List {
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
