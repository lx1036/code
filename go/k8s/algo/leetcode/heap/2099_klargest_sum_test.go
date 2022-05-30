package heap

import (
	"k8s.io/klog/v2"
	"testing"
)

// 重复练习
// https://leetcode.cn/problems/find-subsequence-of-length-k-with-the-largest-sum/solution/you-xian-ji-dui-lie-by-lx1036-80z7/

func maxSubsequence(nums []int, k int) []int {
	pq := NewPriorityQueue2099()
	for key, value := range nums {
		pq.Push(&Item2099{
			key:   key,
			value: value,
		})
	}

	i := 0
	match := map[int]bool{}
	for len(pq.items) != 0 {
		item := pq.Pop()
		match[item.key] = true
		i++
		if i == k {
			break
		}
	}
	var values []int
	for key, value := range nums {
		if match[key] {
			values = append(values, value)
		}
	}

	return values
}

func TestMaxSubsequence(test *testing.T) {
	nums := []int{2, 1, 3, 3}
	k := 2
	klog.Info(maxSubsequence(nums, k))

	nums = []int{-1, -2, 3, 4}
	k = 3
	klog.Info(maxSubsequence(nums, k))

	nums = []int{3, 4, 3, 3}
	k = 2
	klog.Info(maxSubsequence(nums, k))
}

type Item2099 struct {
	key   int
	value int
}

func (item *Item2099) Less(than *Item2099) bool {
	return item.value > than.value // 最大堆
	//return item.value < than.value // 最小堆
}

// 使用最小堆
type PriorityQueue2099 struct {
	items []*Item2099
}

func NewPriorityQueue2099() *PriorityQueue2099 {
	return &PriorityQueue2099{
		items: make([]*Item2099, 0),
	}
}

func (pq *PriorityQueue2099) Push(item *Item2099) {
	pq.items = append(pq.items, item)
	pq.up(len(pq.items) - 1)
}

func (pq *PriorityQueue2099) Pop() *Item2099 {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueue2099) Peek() *Item2099 {
	return pq.items[0]
}

func (pq *PriorityQueue2099) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueue2099) down(index int) {
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

func (pq *PriorityQueue2099) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
