package k8s

import (
	"k8s.io/klog/v2"
	"testing"
)

type LruCache struct {
	capacity int
	list     *List
	items    map[int]*Element
}

type Element struct {
	prev, next *Element
	key, value int
}

type List struct {
	root Element
	len  int
}

func NewList() *List {
	l := &List{}
	l.root.prev = &l.root
	l.root.next = &l.root
	return l
}
func (l *List) MoveFront(e *Element) {
	if l.root.next == e {
		return
	}

	prev := e.prev
	next := e.next
	prev.next = next
	next.prev = prev

	next = l.root.next
	next.prev = e
	e.next = next
	l.root.next = e
	e.prev = &l.root
}
func (l *List) PushFront(e *Element) {
	if l.root.next == e {
		return
	}

	next := l.root.next
	next.prev = e
	e.next = next
	l.root.next = e
	e.prev = &l.root

	l.len++
}
func (l *List) LastElement() *Element {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}
func (l *List) RemoveLastElement(e *Element) {
	e.prev.next = &l.root
	l.root.prev = e.prev
	l.len--
}

func NewLruCache(capacity int) *LruCache {
	return &LruCache{
		list:     NewList(),
		items:    make(map[int]*Element),
		capacity: capacity,
	}
}
func (c *LruCache) Get(key int) int {
	if e, ok := c.items[key]; ok {
		c.list.MoveFront(e)
		return e.value
	}

	return -1
}
func (c *LruCache) Put(key, value int) {
	if e, ok := c.items[key]; ok {
		c.list.MoveFront(e)
		e.value = value
		return
	}

	e := &Element{key: key, value: value}
	c.items[key] = e
	c.list.PushFront(e)
	if c.list.len > c.capacity {
		e = c.list.LastElement()
		if e != nil {
			delete(c.items, e.key)
			c.list.RemoveLastElement(e)
		}
	}
}

func TestLRUCache(test *testing.T) {
	cache := NewLruCache(2)
	cache.Put(1, 1)
	cache.Put(2, 2)
	if cache.Get(1) != 1 {
		klog.Fatal("get 1 error")
	}
	cache.Put(3, 3)
	if cache.Get(2) != -1 {
		klog.Fatal("get 2 error")
	}
	cache.Put(4, 4)
	if cache.Get(1) != -1 {
		klog.Fatal("get 1 error")
	}
	if cache.Get(3) != 3 {
		klog.Fatal("get 3 error")
	}
	if cache.Get(4) != 4 {
		klog.Fatal("get 4 error")
	}
	cache.Put(3, 5)
	if cache.Get(3) != 5 {
		klog.Fatal("get 3 error")
	}
}
