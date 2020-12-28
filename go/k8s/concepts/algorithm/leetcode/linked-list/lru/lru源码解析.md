


# Kubernetes学习笔记之LRU算法源码解析

## Overview
本文章基于k8s release-1.17分支代码。

之前一篇文章学习 **[Kubernetes学习笔记之ServiceAccount TokensController源码解析](https://juejin.cn/post/6910867953265147912)** ，主要学习
ServiceAccount有关知识，发现其中使用了LRU Cache，代码在 **[L106](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/controller/serviceaccount/tokens_controller.go#L106)** 。
k8s自己封装了一个LRU cache的对象 **[MutationCache](https://github.com/kubernetes/kubernetes/blob/release-1.17/staging/src/k8s.io/client-go/tools/cache/mutation_cache.go)** ，
正好趁此机会复习下 LRU 算法知识。


LRU算法一般也是面试必考算法题，算法内容也很简单很直观，主要是通过在固定容量空间内，不常被访问被认为旧数据可以先删除，最近被访问的数据可以认为后面被访问概率很大，作为最新的数据。比如，
**[漫画：什么是LRU算法？](https://zhuanlan.zhihu.com/p/52196637)** 这幅漫画描述的那样，在容量有限情况下，可以删除那些最老的用户数据，留下最新的用户数据。
这样就感觉数据按照倒叙排列似的，最前面的是最新的，最末尾的是最旧的数据。

数据存储可以通过双向链表存储，而不是单向链表，因为当知道链表的一个元素element时，可以通过element.prev和element.next指针就能知道当前元素的前驱元素和
后驱元素，删除和添加操作算法复杂度都是O(1)，而单向链表无法做到这一点。

另外一个问题是如何知道O(1)的查询到一个元素element的值，这可以通过哈希表即 `map[key]*element` 结构知道，只要知道key，就立刻O(1)知道element，
再结合双向链表的O(1)删除和O(1)添加操作。

通过组合双向链表和哈希表组成的一个lru数据结构，就可以实现删除旧数据、读取新数据和插入新数据算法复杂度都是O(1)，这就很厉害很高效的算法了。


## 设计编写LRU算法代码
首先是设计出一个双向链表list，可以直接使用golang自带的双向链表，代码在 /usr/local/go/src/container/list/list.go ，本文这里参考源码写一个并学习之。

首先设计双向链表的结构，Element对象是链表中的节点元素。这里最关键设计是list的占位元素root，是个值为空的元素，其root.next是链表的第一个元素head，
其root.prev是链表的最后一个元素tail，这个设计是直接O(1)知道链表的首位元素，这样链表list就构成了一个链表环ring：

```go
// 算法设计：使用哈希表+双向链表实现
type Element struct {
	prev, next *Element

	Value interface{}
}

// root这个设计很巧妙，连着双向链表的head和tail，可以看Front()和Back()函数
// 获取双向链表的第一个和最后一个元素。root类似一个占位元素
type list struct {
	root Element
	len  int
}

// root是一个empty Element，作为补位元素使得list为一个ring
// list.root.next 是双向链表的第一个元素；list.root.prev 是双向链表的最后一个元素
func (l *list) Init() *list {
    l.root.prev = &l.root
    l.root.next = &l.root
    l.len = 0
    
    return l
}

func (l *list) Len() int {
    return l.len
}
```

然后就是双向链表的新加入一个元素并置于最前面、移动某个元素置于最前面、从链表中删除某个元素这三个重要方法。
新加入一个元素并置于最前面方法，比较简单：

```go
// element置于newest位置，置于最前
func (l *list) PushFront(v interface{}) *Element {
	e := &Element{
		Value: v,
	}

	return l.insert(e, &l.root)
}

// e插入at的位置，at/e/at.next指针需要重新赋值
func (l *list) insert(e, at *Element) *Element {
	// 插入当前位置
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e

	l.len++

	return e
}
```

移动某个元素置于最前面方法：

```go
// 把e置双向链表最前面
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
```

从链表中删除某个元素方法：

```go
func (l *list) Remove(e *Element) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.prev = nil
	e.next = nil

	l.len--
}
```

以上逻辑都比较简单，最后加上返回链表的head和tail元素等等方法：

```go
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

func (l *list) Prev(e *Element) *Element {
	p := e.prev
	if p != &l.root {
		return p
	}

	return nil
}
```

可见设计出这样的一个双向链表还是比较简单的，接下来就是LRU对象了。LRU对象包含双向链表，同时包含哈希表 `map[interface{}]*Element` 来O(1)查询某个
key的Element数据，完整LRU代码如下：

```go

type LRU struct {
	// 指定LRU固定长度，超过的旧数据则移除
	capacity int

	// 双向链表，链表存储每一个*list.Element
	cache *list

	// 哈希表，每一个key是Entry的key
	items map[interface{}]*Element
}

type Entry struct {
	key   interface{}
	value interface{}
}

func NewLRU(capacity int) (*LRU, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacity must be positive")
	}

	cache := &LRU{
		capacity: capacity,
		cache:    new(list).Init(),
		items:    make(map[interface{}]*Element),
	}

	return cache, nil
}

func (c *LRU) Purge() {
	c.cache.Init()
	c.items = make(map[interface{}]*Element)
	c.capacity = 0
}

// 添加一个Entry，O(1)
func (c *LRU) Add(key, value interface{}) (evicted bool) {
	// (key,value)已经存在LRU中
	if element, ok := c.items[key]; ok {
		c.cache.MoveToFront(element)         // 从双向链表中置前，从原有位置删除，然后置最前
		element.Value.(*Entry).value = value // 更新值

		return false
	}

	entry := &Entry{key: key, value: value}
	ent := c.cache.PushFront(entry) // 新元素置最前
	c.items[key] = ent

	evict := c.cache.Len() > c.capacity
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

// 双向链表的长度
func (c *LRU) Len() int {
	return c.cache.Len()
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	// 这里从最末端，即最旧的数据开始查询
	for ent := c.cache.Back(); ent != nil; ent = c.cache.Prev(ent) {
		keys[i] = ent.Value.(*Entry).key
		i++
	}

	return keys
}

func (c *LRU) Remove(key interface{}) bool {
	e, ok := c.items[key]
	if !ok {
		return false
	}

	// 从双向链表中删除element，复杂度O(1)，同时从哈希表items中删除
	c.cache.Remove(e)
	delete(c.items, key)

	return true
}

```


设计好了LRU对象，然后代码测试验证下结果正确性：

```go
// 执行结果没问题
func TestSimpleLRU(test *testing.T) {
	l, _ := NewLRU(128)
	for i := 0; i < 256; i++ {
		l.Add(i, i)
	}
	if l.Len() != 128 {
		panic(fmt.Sprintf("bad len: %v", l.Len()))
	}

	// 这里v==i+128才正确，0-127已经被删除了
	for i, k := range l.Keys() {
		if v, ok := l.Get(k); !ok || v != k || v != i+128 {
			test.Fatalf("bad key: %v", k)
		}
	}

	for i := 0; i < 128; i++ {
		_, ok := l.Get(i)
		if ok {
			test.Fatalf("should be evicted")
		}
	}
	for i := 128; i < 256; i++ {
		_, ok := l.Get(i)
		if !ok {
			test.Fatalf("should not be evicted")
		}
	}

	for i := 128; i < 192; i++ {
		ok := l.Remove(i)
		if !ok {
			test.Fatalf("should be contained")
		}
		ok = l.Remove(i)
		if ok {
			test.Fatalf("should not be contained")
		}
		_, ok = l.Get(i)
		if ok {
			test.Fatalf("should be deleted")
		}
	}
	l.Get(192) // expect 192 to be last key in l.Keys()

	for i, k := range l.Keys() {
		if (i < 63 && k != i+193) || (i == 63 && k != 192) {
			test.Fatalf("out of order key: %v", k)
		}
	}

	l.Purge()
	if l.Len() != 0 {
		test.Fatalf("bad len: %v", l.Len())
	}
	if _, ok := l.Get(200); ok {
		test.Fatalf("should contain nothing")
	}
}
```


**[漫画：什么是LRU算法？](https://zhuanlan.zhihu.com/p/52196637)** 这篇文章中小灰遇到了一个难题，用户系统要爆炸了，不知道怎么去删除那些
缓存的用户数据来减少内存使用，肯定不是随机删除。但是通过双向链表加上哈希表简单组合，构成了一个强大靠谱的LRU结构，删除最旧的数据，保留最新的数据
(这里假设最近被访问的数据是新数据，未被访问的数据则排队置后)，就完美解决了难题，可见LRU算法的巧妙强大。k8s源码中同样使用了LRU结构，
不会LRU算法看k8s源码都费劲。可见算法和数据结构的重要性，刷leetcode是个需要一直坚持下去的活。



## 参考文献

**[漫画：什么是LRU算法？](https://zhuanlan.zhihu.com/p/52196637)**

**[mutation_cache.go使用LRU Cache](https://github.com/kubernetes/kubernetes/blob/release-1.17/staging/src/k8s.io/client-go/tools/cache/mutation_cache.go)**

golang自带双向链表：/usr/local/go/src/container/list/list.go

**[golang-lru](github.com/hashicorp/golang-lru)**

**[leetcode #146](https://leetcode-cn.com/problems/lru-cache/)**

**[leetcode #460](https://leetcode-cn.com/problems/lfu-cache/)**

**[leetcode #1625](https://leetcode-cn.com/problems/lru-cache-lcci/)**
