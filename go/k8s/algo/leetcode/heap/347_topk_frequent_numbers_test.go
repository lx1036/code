package heap

import (
	"testing"

	"k8s.io/klog/v2"
)

// 最小堆，堆排序

// https://leetcode-cn.com/problems/top-k-frequent-elements/

func topKFrequent347(nums []int, k int) []int {
	record := map[int]int{}
	for _, num := range nums {
		if value, ok := record[num]; !ok {
			record[num] = 1
		} else {
			record[num] = value + 1
		}
	}

	pq := NewPriorityQueue347()
	for key, value := range record {
		pq.Push(&Item347{key: key, value: value})
	}

	i := 1
	var results []int
	for len(pq.items) != 0 {
		item := pq.Pop()
		results = append(results, item.key)
		if i == k {
			break
		}
		i++
	}

	return results
}

func TestTopKFrequent347(test *testing.T) {
	nums := []int{1, 1, 1, 2, 2, 3}
	k := 2
	klog.Info(topKFrequent347(nums, k))

	nums = []int{1}
	k = 1
	klog.Info(topKFrequent347(nums, k))
}

type Item347 struct {
	key   int
	value int
}

func (item *Item347) Less(than *Item347) bool {
	return item.value > than.value // 最大堆
	//return item.value < than.value // 最小堆
}

// 使用最小堆
type PriorityQueue347 struct {
	items []*Item347
}

func NewPriorityQueue347() *PriorityQueue347 {
	return &PriorityQueue347{
		items: make([]*Item347, 0),
	}
}

func (pq *PriorityQueue347) Push(item *Item347) {
	pq.items = append(pq.items, item)
	pq.up(len(pq.items) - 1)
}

func (pq *PriorityQueue347) Pop() *Item347 {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueue347) Peek() *Item347 {
	return pq.items[0]
}

func (pq *PriorityQueue347) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueue347) down(index int) {
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

func (pq *PriorityQueue347) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
