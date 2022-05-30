package heap

import (
	"testing"

	"k8s.io/klog/v2"
)

// https://leetcode.cn/problems/kth-largest-element-in-a-stream/solution/you-xian-ji-dui-lie-by-lx1036-e4s6/

type KthLargest struct {
	pq *PriorityQueue703
}

func Constructor703(k int, nums []int) KthLargest {
	l := KthLargest{
		pq: NewPriorityQueue703(k),
	}

	for _, num := range nums {
		l.pq.Push(&Item703{
			value: num,
		})
	}

	return l
}

func (this *KthLargest) Add(val int) int {
	this.pq.Push(&Item703{
		value: val,
	})

	item := this.pq.Peek()
	return item.value
}

func TestKthLargest(test *testing.T) {
	kth := Constructor703(3, []int{4, 5, 8, 2})
	klog.Info(kth.Add(3))  // 4
	klog.Info(kth.Add(5))  // 5
	klog.Info(kth.Add(10)) // 5
	klog.Info(kth.Add(9))  // 8
	klog.Info(kth.Add(4))  // 8
}

type Item703 struct {
	key   int
	value int
}

func (item *Item703) Less(than *Item703) bool {
	//return item.value > than.value // 最大堆
	return item.value < than.value // 最小堆
}

// 使用最小堆
type PriorityQueue703 struct {
	k     int
	items []*Item703
}

func NewPriorityQueue703(k int) *PriorityQueue703 {
	return &PriorityQueue703{
		k:     k,
		items: make([]*Item703, 0),
	}
}

func (pq *PriorityQueue703) Push(item *Item703) {
	pq.items = append(pq.items, item)
	pq.up(len(pq.items) - 1)

	// 保证 k 个元素
	for {
		if len(pq.items) <= pq.k {
			break
		}
		pq.Pop()
	}
}

func (pq *PriorityQueue703) Pop() *Item703 {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueue703) Peek() *Item703 {
	return pq.items[0]
}

func (pq *PriorityQueue703) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueue703) down(index int) {
	for {
		left := 2*index + 1
		l := len(pq.items) - 1
		if left > l {
			break
		}
		right := left + 1
		min := left
		if right <= l && pq.items[right].Less(pq.items[min]) {
			min = right
		}

		if pq.items[index].Less(pq.items[min]) {
			break
		}

		pq.swap(index, min)
		index = min
	}
}

func (pq *PriorityQueue703) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
