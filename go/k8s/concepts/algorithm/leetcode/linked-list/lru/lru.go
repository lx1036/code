package lru

import (
	"fmt"
)

// 算法设计：使用哈希表+双向链表实现

type Element struct {
	prev, next *Element

	Value interface{}
}

type list struct {
	root Element
	len  int
}

func (l *list) Init() *list {
	l.root.prev = &l.root
	l.root.next = &l.root
	l.len = 0

	return l
}

func (l *list) MoveToFront(e *Element) {
	if e == l.root.next {
		return
	}

	l.move(e, &l.root)
}

func (l *list) move(e, at *Element) {
	if e == at {
		return
	}

	// 从原来位置删除
	e.prev.next = e.next
	e.next.prev = e.prev

	// 插入当前位置
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

func (l *list) Len() int {
	return l.len
}

func (l *list) PushFront(v interface{}) *Element {
	e := &Element{
		Value: v,
	}

	return l.insert(e, &l.root)
}

func (l *list) insert(e, at *Element) *Element {
	// 插入当前位置
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e

	l.len++

	return e
}

// 返回list的最后一个元素
func (l *list) Back() *Element {
	if l.len == 0 {
		return nil
	}

	// 这里list是一个ring
	return l.root.prev
}

// 返回list的最前一个元素
func (l *list) Front() *Element {
	if l.len == 0 {
		return nil
	}

	// 这里list是一个ring
	return l.root.next
}

func (l *list) Remove(e *Element) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.prev = nil
	e.next = nil

	l.len--
}

type LRU struct {
	// 指定LRU固定长度，超过的旧数据则移除
	size int

	// 双向链表，链表存储每一个*list.Element
	cache *list

	// 哈希表，每一个key是Entry的key
	items map[interface{}]*Element
}

type Entry struct {
	key   interface{}
	value interface{}
}

func NewLRU(size int) (*LRU, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size must be positive")
	}

	cache := &LRU{
		size:  size,
		cache: new(list).Init(),
		items: make(map[interface{}]*Element),
	}

	return cache, nil
}

func (c *LRU) Purge() {

}

// 添加一个Entry，O(1)
func (c *LRU) Add(key, value interface{}) (evicted bool) {
	// (key,value)已经存在LRU中
	if element, ok := c.items[key]; ok {
		c.cache.MoveToFront(element)         // 从双向链表中置前
		element.Value.(*Entry).value = value // 更新值

		return false
	}

	entry := &Entry{key: key, value: value}
	ent := c.cache.PushFront(entry)
	c.items[key] = ent

	evict := c.cache.Len() > c.size
	if evict {
		// 如果超过指定长度，移除旧数据
		c.removeOldest()
	}

	return evict
}

func (c *LRU) Get(key interface{}) (value interface{}, ok bool) {
	if element, ok := c.items[key]; ok {
		// 置前，复杂度O(1)
		c.cache.MoveToFront(element)
		return element.Value.(*Entry).value, true
	}

	return nil, false
}

// 删除最旧的数据O(1)
func (c *LRU) removeOldest() {
	element := c.cache.Back()
	if element != nil {
		c.removeElement(element)
	}
}

// 算法复杂度O(1)
func (c *LRU) removeElement(element *Element) {
	// 直接使用双向链表的Remove()，复杂度O(1)
	c.cache.Remove(element)
	key := element.Value.(*Entry).key
	// 别忘了从哈希表中删除Entry.key
	delete(c.items, key)
}
