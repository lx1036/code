package daemon

import (
	"time"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/types"
)

// @see https://github.com/kubernetes/kubernetes/blob/v1.23.5/staging/src/k8s.io/client-go/util/workqueue/delaying_queue.go
// @see https://github.com/AliyunContainerService/terway/blob/main/pkg/pool/queue.go

type poolItem struct {
	res           types.NetworkResource
	reservation   time.Time
	idempotentKey string
}

func (item *poolItem) Less(other *poolItem) bool {
	return item.reservation.Before(other.reservation)
}

type PriorityQueue struct {
	items []*poolItem
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: make([]*poolItem, 0),
	}
}

func (pq *PriorityQueue) Push(item *poolItem) {
	pq.items = append(pq.items, item)
	index := len(pq.items)
	pq.up(index - 1)
}

func (pq *PriorityQueue) Peek() *poolItem {
	if len(pq.items) == 0 {
		return nil
	}

	return pq.items[0]
}

func (pq *PriorityQueue) Pop() *poolItem {
	result := pq.items[0]
	length := len(pq.items)
	pq.items[0] = pq.items[length-1]
	pq.items = pq.items[:length-1]

	pq.down(0)
	return result
}

func (pq *PriorityQueue) Size() int {
	return len(pq.items)
}

func (pq *PriorityQueue) up(index int) {
	for {
		parent := (index - 1) / 2                                       // parent==index==0 即最小堆顶端
		if parent == index || !pq.items[index].Less(pq.items[parent]) { // 要么最小堆顶端，要么大于等于父节点
			break
		}
		pq.swap(index, parent)
		index = parent // 继续向上比较
	}
}

func (pq *PriorityQueue) down(index int) {
	for {
		left := 2*index + 1
		if left >= len(pq.items) {
			break // index 已经是叶子节点
		}
		minChild := left
		right := left + 1
		if right < len(pq.items) && pq.items[right].Less(pq.items[left]) { // 右边有节点且小于左节点
			minChild = right
		}
		if pq.items[index].Less(pq.items[minChild]) {
			break // index 父节点比右节点还小，即父节点就是最小节点
		}

		pq.swap(index, minChild)
		index = minChild // 继续向下比较
	}
}

func (pq *PriorityQueue) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
