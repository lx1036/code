package heap

import (
	"testing"

	"k8s.io/klog/v2"
)

// https://leetcode.cn/problems/ugly-number-ii/solution/zui-xiao-dui-by-lx1036-3iid/

// 1 是丑数并放入最小堆，每次 x=Pop()，再放入 2x,3x,5x 这些丑数。同时使用 map 去重

var factors = []int{2, 3, 5}

func nthUglyNumber(n int) int {
	pq := NewPriorityQueue264()
	pq.Push(&Item264{value: 1})

	var result int
	results := map[int]bool{1: true}
	for i := 1; ; i++ {
		item := pq.Pop()
		if i == n {
			result = item.value
			break
		}

		for _, factor := range factors {
			data := factor * item.value
			if _, ok := results[data]; !ok { // 去重再push
				results[data] = true
				pq.Push(&Item264{value: data})
			}
		}
	}

	return result
}

func TestNthUglyNumber(test *testing.T) {
	klog.Info(nthUglyNumber(10))
	klog.Info(nthUglyNumber(11))
	klog.Info(nthUglyNumber(1))
	klog.Info(nthUglyNumber(299))
	klog.Info(nthUglyNumber(300))
}

type Item264 struct {
	value int
}

func (item *Item264) Less(than *Item264) bool {
	//return item.value > than.value // 最大堆
	return item.value < than.value // 最小堆
}

// 使用最小堆
type PriorityQueue264 struct {
	items []*Item264
}

func NewPriorityQueue264() *PriorityQueue264 {
	return &PriorityQueue264{
		items: make([]*Item264, 0),
	}
}

func (pq *PriorityQueue264) Push(item *Item264) {
	pq.items = append(pq.items, item)
	pq.up(len(pq.items) - 1)
}

func (pq *PriorityQueue264) Pop() *Item264 {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueue264) Peek() *Item264 {
	return pq.items[0]
}

func (pq *PriorityQueue264) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueue264) down(index int) {
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

func (pq *PriorityQueue264) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
