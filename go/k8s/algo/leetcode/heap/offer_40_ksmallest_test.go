package heap

import (
	"testing"

	"k8s.io/klog/v2"
)

// 优先队列 priority queue

// https://leetcode-cn.com/problems/zui-xiao-de-kge-shu-lcof/

func getLeastNumbers(arr []int, k int) []int {
	if k == 0 {
		return nil
	}

	pq := NewPriorityQueueOffer40()
	for _, value := range arr {
		pq.Push(&ItemOffer40{value: value})
	}

	var results []int
	i := 1
	for len(pq.items) != 0 {
		results = append(results, pq.Pop().value)
		if i == k {
			break
		}
		i++
	}

	return results
}

func TestGetLeastNumbers(test *testing.T) {
	values := []int{3, 2, 1}
	k := 2
	klog.Info(getLeastNumbers(values, k))

	values = []int{0, 1, 2, 1}
	k = 1
	klog.Info(getLeastNumbers(values, k))
}

type ItemOffer40 struct {
	key   int
	value int
}

func (item *ItemOffer40) Less(than *ItemOffer40) bool {
	//return item.value > than.value // 最大堆
	return item.value <= than.value // 最小堆
}

// 使用最小堆
type PriorityQueueOffer40 struct {
	items []*ItemOffer40
}

func NewPriorityQueueOffer40() *PriorityQueueOffer40 {
	return &PriorityQueueOffer40{
		items: make([]*ItemOffer40, 0),
	}
}

func (pq *PriorityQueueOffer40) Push(item *ItemOffer40) {
	pq.items = append(pq.items, item)
	pq.up(len(pq.items) - 1)
}

func (pq *PriorityQueueOffer40) Pop() *ItemOffer40 {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueueOffer40) Peek() *ItemOffer40 {
	return pq.items[0]
}

func (pq *PriorityQueueOffer40) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueueOffer40) down(index int) {
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

func (pq *PriorityQueueOffer40) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
