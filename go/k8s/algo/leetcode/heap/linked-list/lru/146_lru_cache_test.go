package lru

// LRU: Least Recently Used
// https://leetcode-cn.com/problems/lru-cache/

// 阿里面试题: https://zhuanlan.zhihu.com/p/76708575

// 字节面试题：**[LRU原理和Redis实现——一个今日头条的面试题](https://zhuanlan.zhihu.com/p/34133067)**

type entry struct {
	key, value int
}

type LRUCache struct {
	capacity int

	cache *list

	items map[int]*Element
}

func Constructor(capacity int) LRUCache {
	c := LRUCache{
		capacity: capacity,
		cache:    new(list).Init(),
		items:    make(map[int]*Element),
	}

	return c
}

func (cache *LRUCache) Get(key int) int {
	if element, ok := cache.items[key]; ok {
		return element.Value.(*entry).value
	}

	return 0
}

func (cache *LRUCache) Put(key int, value int) {
	if element, ok := cache.items[key]; ok {
		cache.cache.MoveToFront(element)
		element.Value.(*entry).value = value
		return
	}

	e := &entry{key, value}

	cache.cache.PushFront(e)
	if cache.cache.Len() > cache.capacity {
		cache.removeOldest()
	}
}

func (cache *LRUCache) removeOldest() {
	e := cache.cache.Back()
	if e != nil {
		cache.removeElement(e)
	}
}

func (cache *LRUCache) removeElement(e *Element) {
	cache.cache.Remove(e)
	key := e.Value.(*entry).key
	delete(cache.items, key)
}

/**
 * Your LRUCache object will be instantiated and called as such:
 * obj := Constructor(capacity);
 * param_1 := obj.Get(key);
 * obj.Put(key,value);
 */
