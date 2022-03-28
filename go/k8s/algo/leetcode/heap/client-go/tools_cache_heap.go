package client_go

// client-go/tools/cache/heap.go

type heapItem struct {
	index int // heap queue 中的索引

	obj interface{} // 具体的数据值
}

type LessFunc func(interface{}, interface{}) bool

// 哈希表+queue
type heapData struct {
	// key为push进来的itemKeyValue.key
	items map[string]*heapItem

	queue []string

	// lessFunc is used to compare two objects in the heap.
	lessFunc LessFunc
}

func (h *heapData) Len() int {
	return len(h.queue)
}

func (h *heapData) Less(i, j int) bool {
	if i > len(h.queue) || j > len(h.queue) {
		return false
	}

	itemi, ok := h.items[h.queue[i]]
	if !ok {
		return false
	}
	itemj, ok := h.items[h.queue[j]]
	if !ok {
		return false
	}

	return h.lessFunc(itemi.obj, itemj.obj)
}

func (h *heapData) Swap(i, j int) {

}

type itemKeyValue struct {
	key string // 哈希表中的key
	obj interface{}
}

func (h *heapData) Push(kv interface{}) {
	keyValue := kv.(*itemKeyValue)

	n := len(h.queue)
	h.items[keyValue.key] = &heapItem{obj: keyValue.obj, index: n}

	h.queue = append(h.queue, keyValue.key)
}

func (h *heapData) Pop() interface{} {

}
