package lru

// LRU: Least Recently Used
// https://leetcode-cn.com/problems/lru-cache/

// 阿里面试题: https://zhuanlan.zhihu.com/p/76708575

// 字节面试题：**[LRU原理和Redis实现——一个今日头条的面试题](https://zhuanlan.zhihu.com/p/34133067)**

type LRUCache struct {
}

func Constructor(capacity int) LRUCache {

	return LRUCache{}
}

func (cache *LRUCache) Get(key int) int {

	return 0
}

func (cache *LRUCache) Put(key int, value int) {

}

/**
 * Your LRUCache object will be instantiated and called as such:
 * obj := Constructor(capacity);
 * param_1 := obj.Get(key);
 * obj.Put(key,value);
 */
