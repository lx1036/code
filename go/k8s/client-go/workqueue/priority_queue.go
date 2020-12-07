package workqueue

import "time"

type t interface {}

// 包装添加的数据，加上时间
type waitFor struct {
	data t
	readyAt time.Time // 决定priority

	// priority queue(heap)中的索引，即[]*waitFor数组中的索引index
	index int
}

// usr/local/go/src/container/heap/heap.go
type waitForPriorityQueue []*waitFor

func (pq waitForPriorityQueue) Len() int {
	return len(pq)
}

func (pq waitForPriorityQueue) Less(i, j int) bool {
	return pq[i].readyAt.Before(pq[j].readyAt)
}

func (pq waitForPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq waitForPriorityQueue) Peek() interface{} {
	return pq[0]
}

func (pq *waitForPriorityQueue) Push(x interface{})  {
	n := len(*pq)
	item := x.(*waitFor)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *waitForPriorityQueue) Pop() interface{} {
	n := len(*pq)
	item := (*pq)[n-1]
	item.index = -1
	*pq = (*pq)[0:(n - 1)]
	return item
}


