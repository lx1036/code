package heap

import (
	"container/heap"
	"fmt"
	"testing"
)

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

func TestPriorityQueue(test *testing.T) {
	items := map[string]int{
		"golang":     8,
		"php":        4,
		"kubernetes": 10,
		"mysql":      6,
		"redis":      2,
		"algo":       9,
		"kafka":      6,
	}

	i := 0
	//pq := make(PriorityQueue, len(items))
	var pq PriorityQueue
	for value, priority := range items {
		pq = append(pq, &item{
			value:    value,
			priority: priority,
			index:    i,
		})

		i++
	}

	// 初始化为priority-queue
	heap.Init(&pq)
	obj := &item{
		value:    "prometheus",
		priority: 7,
	}
	heap.Push(&pq, obj)
	obj.value = "prometheus operator"
	obj.priority = 11
	heap.Fix(&pq, obj.index)
	for pq.Len() > 0 {
		o := heap.Pop(&pq).(*item)
		fmt.Println(fmt.Sprintf("%s/%d", o.value, o.priority))
	}
}
