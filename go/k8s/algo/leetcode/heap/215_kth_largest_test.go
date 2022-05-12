package heap

import (
	"testing"

	"k8s.io/klog/v2"
)

// https://leetcode-cn.com/problems/kth-largest-element-in-an-array/

// 堆排序/最小堆

func findKthLargest(nums []int, k int) int {
	pq := NewPriorityQueue215(k)
	for _, num := range nums {
		pq.Push(&Item215{value: num})
	}

	return pq.Peek().value
}

func TestFindKthLargest(test *testing.T) {
	nums := []int{3, 2, 1, 5, 6, 4}
	k := 2
	klog.Info(findKthLargest(nums, k))

	nums = []int{3, 2, 3, 1, 2, 4, 5, 5, 6}
	k = 4
	klog.Info(findKthLargest(nums, k))
}

type Item215 struct {
	key   int
	value int
}

func (item *Item215) Less(than *Item215) bool {
	//return item.value > than.value // 最大堆
	return item.value <= than.value // 最小堆
}

// 使用最小堆
type PriorityQueue215 struct {
	k     int
	items []*Item215
}

func NewPriorityQueue215(k int) *PriorityQueue215 {
	return &PriorityQueue215{
		k:     k,
		items: make([]*Item215, 0),
	}
}

func (pq *PriorityQueue215) Push(item *Item215) {
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

func (pq *PriorityQueue215) Pop() *Item215 {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueue215) Peek() *Item215 {
	return pq.items[0]
}

func (pq *PriorityQueue215) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueue215) down(index int) {
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

func (pq *PriorityQueue215) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
