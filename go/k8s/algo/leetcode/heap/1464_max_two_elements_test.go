package heap

import (
	"k8s.io/klog/v2"
	"testing"
)

// https://leetcode.cn/problems/maximum-product-of-two-elements-in-an-array/

func maxProduct(nums []int) int {
	pq := NewPriorityQueue1464()
	for _, num := range nums {
		pq.Push(&Item1464{value: num})
	}

	r1 := pq.Pop()
	r2 := pq.Pop()

	return (r1.value - 1) * (r2.value - 1)
}

func TestMaxProduct(test *testing.T) {
	nums := []int{3, 4, 5, 2}
	klog.Info(maxProduct(nums))

	nums = []int{1, 5, 4, 5}
	klog.Info(maxProduct(nums))

	nums = []int{3, 7}
	klog.Info(maxProduct(nums))
}

type Item1464 struct {
	key   int
	value int
}

func (item *Item1464) Less(than *Item1464) bool {
	return item.value > than.value // 最大堆
	//return item.value < than.value // 最小堆
}

// 使用最小堆
type PriorityQueue1464 struct {
	items []*Item1464
}

func NewPriorityQueue1464() *PriorityQueue1464 {
	return &PriorityQueue1464{
		items: make([]*Item1464, 0),
	}
}

func (pq *PriorityQueue1464) Push(item *Item1464) {
	pq.items = append(pq.items, item)
	pq.up(len(pq.items) - 1)
}

func (pq *PriorityQueue1464) Pop() *Item1464 {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueue1464) Peek() *Item1464 {
	return pq.items[0]
}

func (pq *PriorityQueue1464) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueue1464) down(index int) {
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

func (pq *PriorityQueue1464) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
