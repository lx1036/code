package lru

import (
	"k8s.io/klog/v2"
	"testing"
)

// https://leetcode-cn.com/problems/lfu-cache/

type Element460 struct {
	prev, next *Element460
	key, value int
}

type DoubleList460 struct {
	len  int
	root Element460 // INFO: 不能是指针，这样 DoubleList460 实例化后是一个对象。起到占位作用
}

func NewDoubleList460() *DoubleList460 {
	l := &DoubleList460{
		len: 0,
	}
	l.root.next = &l.root
	l.root.prev = &l.root
	return l
}

func (list *DoubleList460) MoveToFront(element *Element460) {
	if list.root.next == element {
		return
	}

	// 从原来位置删除
	element.prev.next = element.next
	element.next.prev = element.prev

	// 插入 root <=> list[1] 两者中间
	element.prev = &list.root
	element.next = list.root.next
	list.root.next.prev = element
	list.root.next = element
}

func (list *DoubleList460) PushFront(element *Element460) {
	if list.root.next == element {
		return
	}

	// 插入 root <=> list[1] 两者中间
	element.prev = &list.root
	element.next = list.root.next
	list.root.next.prev = element
	list.root.next = element
	list.len++
}

func (list *DoubleList460) Len() int {
	return list.len
}

func (list *DoubleList460) LastElement() *Element460 {
	if list.len == 0 {
		return nil
	}

	return list.root.prev
}

func (list *DoubleList460) RemoveLast(element *Element460) {
	// 从原来位置删除
	element.prev.next = &list.root
	list.root.prev = element.prev
	list.len--
}

type LRUCache460 struct {
	capacity int

	list *DoubleList460

	items map[int]*Element460
}

func Constructor460(capacity int) LRUCache460 {
	c := LRUCache460{
		capacity: capacity,
		list:     NewDoubleList460(),
		items:    make(map[int]*Element460),
	}

	return c
}

func (cache *LRUCache460) Get(key int) int {
	if element, ok := cache.items[key]; ok {
		cache.list.MoveToFront(element)
		return element.value
	}

	return -1
}

func (cache *LRUCache460) Put(key int, value int) {
	if element, ok := cache.items[key]; ok {
		cache.list.MoveToFront(element)
		element.value = value
		return
	}

	e := &Element460{key: key, value: value}
	cache.items[key] = e
	cache.list.PushFront(e)
	if cache.list.Len() > cache.capacity {
		e = cache.list.LastElement()
		if e != nil {
			delete(cache.items, e.key)
			cache.list.RemoveLast(e)
		}
	}
}

//["LFUCache","put","put","get","get","get","put","put","get","get","get","get"]
//[[3],[2,2],[1,1],[2],[1],[2],[3,3],[4,4],[3],[2],[1],[4]]
//[null,null,null,2,1,2,null,null,-1,2,1,4]

func TestLRUCache460(test *testing.T) {
	cache := Constructor(3)
	cache.Put(2, 2)
	cache.Put(1, 1)
	klog.Info(cache.Get(2))
	klog.Info(cache.Get(1))
	klog.Info(cache.Get(2))
	cache.Put(3, 3)
	cache.Put(4, 4)
	klog.Info(cache.Get(3))
	klog.Info(cache.Get(2))
	klog.Info(cache.Get(1))
	klog.Info(cache.Get(4))
}
