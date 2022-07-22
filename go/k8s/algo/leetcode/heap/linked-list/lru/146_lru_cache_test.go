package lru

// LRU: Least Recently Used
// https://leetcode-cn.com/problems/lru-cache/

// 阿里面试题: https://zhuanlan.zhihu.com/p/76708575

// 字节面试题：**[LRU原理和Redis实现——一个今日头条的面试题](https://zhuanlan.zhihu.com/p/34133067)**

import (
	"k8s.io/klog/v2"
	"testing"
)

type Element146 struct {
	prev, next *Element146
	key, value int
}

type DoubleList struct {
	len  int
	root Element146 // INFO: 不能是指针，这样 DoubleList 实例化后是一个对象。起到占位作用
}

func NewDoubleList() *DoubleList {
	l := &DoubleList{
		len: 0,
	}
	l.root.next = &l.root
	l.root.prev = &l.root
	return l
}

func (list *DoubleList) MoveToFront(element *Element146) {
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

func (list *DoubleList) PushFront(element *Element146) {
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

func (list *DoubleList) Len() int {
	return list.len
}

func (list *DoubleList) LastElement() *Element146 {
	if list.len == 0 {
		return nil
	}

	return list.root.prev
}

func (list *DoubleList) RemoveLast(element *Element146) {
	// 从原来位置删除
	element.prev.next = &list.root
	list.root.prev = element.prev
	list.len--
}

type LRUCache146 struct {
	capacity int

	list *DoubleList

	items map[int]*Element146
}

func Constructor(capacity int) LRUCache146 {
	c := LRUCache146{
		capacity: capacity,
		list:     NewDoubleList(),
		items:    make(map[int]*Element146),
	}

	return c
}

func (cache *LRUCache146) Get(key int) int {
	if element, ok := cache.items[key]; ok {
		cache.list.MoveToFront(element)
		return element.value
	}

	return -1
}

func (cache *LRUCache146) Put(key int, value int) {
	if element, ok := cache.items[key]; ok {
		cache.list.MoveToFront(element)
		element.value = value
		return
	}

	e := &Element146{key: key, value: value}
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

func TestLRUCache146(test *testing.T) {
	cache := Constructor(2)
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
