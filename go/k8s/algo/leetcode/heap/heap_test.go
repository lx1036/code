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

//////////////////////////////////////////////////////////////////
// https://www.cnblogs.com/yahuian/p/11945144.html

type Heap []int

// Swap只是交换内容，可以不需要指针
// Push/Pop/Remove不仅仅修改内容，还会修改长度，需要pointer receiver
func (h Heap) Swap(i, j int) {
    h[i], h[j] = h[j], h[i]
}

func (h Heap) Less(i, j int) bool {
    return h[i] < h[j]
}

func (h Heap) up(i int) {
    for {
        parent := (i - 1) / 2
        // 要么i==parent就是达到了顶端
        // 要么i<parent，往上升
        if i == parent || !h.Less(i, parent) {
            break
        }

        h.Swap(i, parent)
        i = parent
    }
}

// Push and Pop use pointer receivers because they modify the slice's length,
// not just its contents.
// 注意go中所有参数转递都是值传递
// 所以要让h的变化在函数外也起作用，此处得传指针
func (h *Heap) Push(x int) {
    *h = append(*h, x)
    h.up(len(*h) - 1)
}

func (h Heap) down(i int) {
    for {
        left := 2*i + 1
        if left >= len(h) {
            break // 已经是最后一个元素了，叶子节点
        }

        // min(left, right)
        child := left
        if right := left + 1; right < len(h) && h.Less(right, child) { // 别忘了right < len(h)条件
            child = right
        }

        if h.Less(i, child) {
            break
        }
        h.Swap(i, child)
        i = child
    }
}

func (h *Heap) Remove(i int) (int, bool) {
    if i < 0 || i >= len(*h) {
        return 0, false
    }

    // 待删除元素与最后元素交换，然后删除
    n := len(*h) - 1
    h.Swap(i, n)
    value := (*h)[n]
    *h = (*h)[0:n]

    if (*h)[i] < (*h)[(i-1)/2] { // 当前节点小于父节点
        h.up(i)
    } else {
        h.down(i)
    }

    return value, true
}

func (h *Heap) Pop() int {
    i := 0

    // 待删除元素与最后元素交换，然后删除
    n := len(*h) - 1
    h.Swap(i, n)
    value := (*h)[n]
    *h = (*h)[0:n]

    h.down(i)

    return value
}

func (h Heap) Init() {
    n := len(h)
    // i > n/2-1 的结点为叶子结点本身已经是堆了
    // 只需要一半元素，另一半会是叶子节点
    for i := n/2 - 1; i >= 0; i-- {
        h.down(i)
    }
}

func TestHeap(test *testing.T) {
    h := Heap{20, 7, 3, 10, 15, 25, 30, 17, 19}
    h.Init()
    fmt.Println(h)

    h.Push(6)
    fmt.Println(h)

    h.Remove(4)
    fmt.Println(h)

    for len(h) > 0 {
        fmt.Println(h.Pop())
    }
}
