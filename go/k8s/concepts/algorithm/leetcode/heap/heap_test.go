package heap

import "sort"

type Interface interface {
	sort.Interface

	Push(x interface{})
	Pop() interface{}
}

// Priority Queue
type item struct {
	value string

	// 根据该值来排序
	priority int
	// 在队列中的索引
	index int
}

type PriorityQueue []*item

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push会修改队列数据，使用指针作为函数receiver
func (pq *PriorityQueue) Push(x interface{}) {
	i := x.(*item)
	i.index = len(*pq)
	*pq = append(*pq, i)
}

func (pq *PriorityQueue) Pop() interface{} {
	n := len(*pq)
	i := (*pq)[n-1]
	i.index = -1 // for safety
	*pq = (*pq)[0:(n - 1)]

	return i
}
