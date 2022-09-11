package k8s

// priority queue 也就是 heap

type PriorityQueue struct {
	items []*Item
}
type Item struct {
	key, value int
}

func (item *Item) Less(other *Item) bool {
	return item.value < other.value
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: make([]*Item, 0),
	}
}

func (q *PriorityQueue) Push(item *Item) {
	q.items = append(q.items, item)
	q.up(len(q.items) - 1)
}

func (q *PriorityQueue) Pop() *Item {
	result := q.items[0]
	last := len(q.items) - 1
	q.items[0] = q.items[last]
	q.items = q.items[:last]
	q.down(0)
	return result
}

func (q *PriorityQueue) up(index int) {
	for {
		parent := (index - 1) / 2
		if parent == index || !q.items[index].Less(q.items[parent]) {
			break
		}
		q.swap(index, parent)
		index = parent
	}
}

func (q *PriorityQueue) swap(index, parent int) {
	q.items[index], q.items[parent] = q.items[parent], q.items[index]
}

func (q *PriorityQueue) down(index int) {
	for {
		left := 2*index + 1
		last := len(q.items) - 1
		if left > last {
			break
		}
		right := left + 1
		min := left
		if right <= last && q.items[right].Less(q.items[min]) {
			min = right
		}

		if q.items[index].Less(q.items[min]) {
			break
		}

		q.swap(index, min)
		index = min
	}
}
