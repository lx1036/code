package heap

import (
	"testing"

	"k8s.io/klog/v2"
)

// https://leetcode.cn/problems/last-stone-weight/solution/you-xian-ji-dui-lie-by-lx1036-2nqv/

func lastStoneWeight(stones []int) int {
	pq := NewPriorityQueue1046()
	for _, stone := range stones {
		pq.Push(&Item1046{value: stone})
	}

	for len(pq.items) != 0 {
		one := pq.Pop()
		if len(pq.items) == 0 {
			return one.value
		}
		two := pq.Pop()
		if one.value-two.value > 0 {
			pq.Push(&Item1046{value: one.value - two.value})
		}
	}

	return 0
}

func TestLastStoneWeight(test *testing.T) {
	klog.Info(lastStoneWeight([]int{2, 7, 4, 1, 8, 1}))
}

type Item1046 struct {
	value int
}

func (item *Item1046) Less(than *Item1046) bool {
	return item.value > than.value // 最大堆
	//return item.value <= than.value // 最小堆
}

type PriorityQueue1046 struct {
	items []*Item1046
}

func NewPriorityQueue1046() *PriorityQueue1046 {
	return &PriorityQueue1046{
		items: make([]*Item1046, 0),
	}
}

func (pq *PriorityQueue1046) Push(item *Item1046) {
	pq.items = append(pq.items, item)
	pq.up(len(pq.items) - 1)

}

func (pq *PriorityQueue1046) Pop() *Item1046 {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueue1046) Peek() *Item1046 {
	return pq.items[0]
}

func (pq *PriorityQueue1046) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueue1046) down(index int) {
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

func (pq *PriorityQueue1046) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
