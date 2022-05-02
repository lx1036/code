package daemon

import (
	"time"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/types"
)

// @see https://github.com/kubernetes/kubernetes/blob/v1.23.5/staging/src/k8s.io/client-go/util/workqueue/delaying_queue.go
// @see https://github.com/AliyunContainerService/terway/blob/main/pkg/pool/queue.go

type eniIPItem struct {
	res        *types.ENIIP
	expiration time.Time
	key        string
}

// 最小/大堆实现 priority queue
type eniIPPriorityQueue []*eniIPItem

func (q *eniIPPriorityQueue) Less(i, j int) bool {
	return q.items[i].reservation.Before(q.items[j].reservation)
}

func (q *eniIPPriorityQueue) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func (q *eniIPPriorityQueue) Push(item *poolItem) {
	q.items = append(q.items, item)

	// bubble up
	index := len(q.items) - 1
	for index > 0 {
		parent := (index - 1) / 2
		if !q.items[index].less(q.items[parent]) {
			break
		}
		q.Swap(index, parent)
		index = parent
	}

}

func (q *eniIPPriorityQueue) Peek() *poolItem {
	return q.items[0]
}

func (q *eniIPPriorityQueue) Pop() *poolItem {
	if q.size == 0 {
		return nil
	}

	item := q.items[0]
	q.items[0] = q.items[q.size-1]
	q.size--
	q.bubbleDown(0)
	return item
}

func (q *eniIPPriorityQueue) Size() int {
	return len(*q)
}
