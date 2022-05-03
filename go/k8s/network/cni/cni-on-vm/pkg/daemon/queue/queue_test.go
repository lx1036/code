package queue

import (
	"fmt"
	"k8s.io/klog/v2"
	"testing"
	"time"
)

// @see Min Heap: https://www.cs.usfca.edu/~galles/visualization/Heap.html
// https://www.cnblogs.com/yahuian/p/11945144.html
// 最小/大堆比较简单

type Item struct {
	value  int
	expire time.Time
}

func (item *Item) Less(other *Item) bool {
	//return item.expire.Before(i.expire)
	return item.value < other.value
}

type PriorityQueue struct {
	items []*Item
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: make([]*Item, 0),
	}
}

func (pq *PriorityQueue) Push(item *Item) {
	pq.items = append(pq.items, item)
	index := len(pq.items)
	pq.up(index - 1)
}

func (pq *PriorityQueue) Peek() *Item {
	if len(pq.items) == 0 {
		return nil
	}

	return pq.items[0]
}

func (pq *PriorityQueue) Pop() *Item {
	result := pq.items[0]
	length := len(pq.items)
	pq.items[0] = pq.items[length-1]
	pq.items = pq.items[:length-1]

	pq.down(0)
	return result
}

func (pq *PriorityQueue) up(index int) {
	for {
		parent := (index - 1) / 2                                       // parent==index==0 即最小堆顶端
		if parent == index || !pq.items[index].Less(pq.items[parent]) { // 要么最小堆顶端，要么大于等于父节点
			break
		}
		pq.swap(index, parent)
		index = parent // 继续向上比较
	}
}

func (pq *PriorityQueue) down(index int) {
	for {
		left := 2*index + 1
		if left >= len(pq.items) {
			break // index 已经是叶子节点
		}
		minChild := left
		right := left + 1
		if right < len(pq.items) && pq.items[right].Less(pq.items[left]) { // 右边有节点且小于左节点
			minChild = right
		}
		if pq.items[index].Less(pq.items[minChild]) {
			break // index 父节点比右节点还小，即父节点就是最小节点
		}

		pq.swap(index, minChild)
		index = minChild // 继续向下比较
	}
}

func (pq *PriorityQueue) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}

func TestPriorityQueue(test *testing.T) {
	pq := NewPriorityQueue()
	data := []*Item{ // 完全二叉树：3 6 15 10 7 20 30 17 19
		&Item{value: 3},
		&Item{value: 10},
		&Item{value: 15},
		&Item{value: 20},
		&Item{value: 30},
		&Item{value: 19},
		&Item{value: 17},
		&Item{value: 6},
		&Item{value: 7},
	}
	for _, value := range data { // 3 10 15 20 30 19 17 6 7
		pq.Push(value)
	}
	for len(pq.items) != 0 {
		item := pq.Pop()
		klog.Infof(fmt.Sprintf("%d", item.value)) // 3 6 7 10 15 17 19 20 30
	}
}

func TestQueue(test *testing.T) {
	pq := NewPriorityQueue()
	items := []*Item{
		&Item{
			value: 1,
		},
		&Item{
			value: 2,
		},
		&Item{
			value: 3,
		},
	}
	pq.items = items
	length := len(pq.items)
	pq.items = pq.items[:length-1]
	for _, item := range pq.items {
		klog.Infof(fmt.Sprintf("%d", item.value))
	}
}
