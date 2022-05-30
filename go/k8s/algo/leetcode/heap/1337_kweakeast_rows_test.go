package heap

import (
	"k8s.io/klog/v2"
	"testing"
)

// https://leetcode.cn/problems/the-k-weakest-rows-in-a-matrix/solution/you-xian-ji-dui-lie-by-lx1036-roiw/

func kWeakestRows(mat [][]int, k int) []int {
	pq := NewPriorityQueue1337()
	for i, row := range mat {
		rowValue := 0
		for _, value := range row {
			rowValue += value
		}
		pq.Push(&Item1337{
			key:   i,
			value: rowValue,
		})
	}

	var results []int
	i := 1
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

func TestWeakestRows(test *testing.T) {
	mat := [][]int{
		{1, 1, 0, 0, 0},
		{1, 1, 1, 1, 0},
		{1, 0, 0, 0, 0},
		{1, 1, 0, 0, 0},
		{1, 1, 1, 1, 1},
	}
	klog.Info(kWeakestRows(mat, 3))

	mat = [][]int{
		{1, 0, 0, 0},
		{1, 1, 1, 1},
		{1, 0, 0, 0},
		{1, 0, 0, 0},
	}
	klog.Info(kWeakestRows(mat, 2))
}

type Item1337 struct {
	key   int
	value int
}

func (item *Item1337) Less(than *Item1337) bool {
	//return item.value > than.value // 最大堆
	return item.value < than.value || (item.value == than.value && item.key < than.key) // 最小堆
}

// 使用最小堆
type PriorityQueue1337 struct {
	items []*Item1337
}

func NewPriorityQueue1337() *PriorityQueue1337 {
	return &PriorityQueue1337{
		items: make([]*Item1337, 0),
	}
}

func (pq *PriorityQueue1337) Push(item *Item1337) {
	pq.items = append(pq.items, item)
	pq.up(len(pq.items) - 1)
}

func (pq *PriorityQueue1337) Pop() *Item1337 {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueue1337) Peek() *Item1337 {
	return pq.items[0]
}

func (pq *PriorityQueue1337) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueue1337) down(index int) {
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

func (pq *PriorityQueue1337) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
