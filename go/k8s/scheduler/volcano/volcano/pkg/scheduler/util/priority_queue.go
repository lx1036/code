package util

import (
	"container/heap"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"
)

// PriorityQueue implements a scheduling queue.
type PriorityQueue struct {
	queue priorityQueue
}

type priorityQueue struct {
	items  []interface{}
	lessFn api.LessFn
}

// NewPriorityQueue returns a PriorityQueue
func NewPriorityQueue(lessFn api.LessFn) *PriorityQueue {
	return &PriorityQueue{
		queue: priorityQueue{
			items:  make([]interface{}, 0),
			lessFn: lessFn,
		},
	}
}

func (q *PriorityQueue) Empty() bool {
	return q.queue.Len() == 0
}

func (q *PriorityQueue) Len() int {
	return q.queue.Len()
}

func (pq *priorityQueue) Len() int {
	return len(pq.items)
}

func (q *PriorityQueue) Pop() interface{} {
	if q.Len() == 0 {
		return nil
	}

	return heap.Pop(&q.queue)
}

func (q *PriorityQueue) Push(it interface{}) {
	heap.Push(&q.queue, it)
}
