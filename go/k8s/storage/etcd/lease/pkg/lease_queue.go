package pkg

import "container/heap"

// LeaseWithTime contains lease object with a time.
type LeaseWithTime struct {
	id LeaseID
	// Unix nanos timestamp.
	time  int64
	index int
}

// PriorityQueue
type LeaseQueue []*LeaseWithTime

func (pq LeaseQueue) Len() int { return len(pq) }

func (pq LeaseQueue) Less(i, j int) bool {
	return pq[i].time < pq[j].time
}

func (pq LeaseQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *LeaseQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*LeaseWithTime)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *LeaseQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

type LeaseExpiredNotifier struct {
	m     map[LeaseID]*LeaseWithTime
	queue LeaseQueue
}

func (mq *LeaseExpiredNotifier) Init() {
	heap.Init(&mq.queue)
	mq.m = make(map[LeaseID]*LeaseWithTime)
	for _, item := range mq.queue {
		mq.m[item.id] = item
	}
}

func (mq *LeaseExpiredNotifier) Len() int {
	return len(mq.m)
}

func (mq *LeaseExpiredNotifier) Poll() *LeaseWithTime {
	if mq.Len() == 0 {
		return nil
	}
	return mq.queue[0]
}

func (mq *LeaseExpiredNotifier) Unregister() *LeaseWithTime {
	item := heap.Pop(&mq.queue).(*LeaseWithTime)
	delete(mq.m, item.id)
	return item
}

func (mq *LeaseExpiredNotifier) RegisterOrUpdate(item *LeaseWithTime) {
	if old, ok := mq.m[item.id]; ok {
		old.time = item.time
		heap.Fix(&mq.queue, old.index)
	} else {
		heap.Push(&mq.queue, item)
		mq.m[item.id] = item
	}
}

func newLeaseExpiredNotifier() *LeaseExpiredNotifier {
	return &LeaseExpiredNotifier{
		m:     make(map[LeaseID]*LeaseWithTime),
		queue: make(LeaseQueue, 0),
	}
}
