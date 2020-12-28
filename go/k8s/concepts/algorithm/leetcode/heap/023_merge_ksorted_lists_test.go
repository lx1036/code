package heap

import (
	"fmt"
	"sort"
	"testing"
)

// 优先队列

// https://leetcode-cn.com/problems/merge-k-sorted-lists/

// 单向链表(没有双向链表好用)
type Element struct {
	next *Element

	Value interface{}
}

type List struct {
	root Element
	len  int
}

func (l *List) Init() *List {
	l.root.next = &l.root
	l.len = 0

	return l
}

func (l *List) AddAt(value interface{}, at *Element) {
	node := &Element{
		Value: value,
	}

	node.next = at.next
	at.next = node

	l.len++
}

func (l *List) AddAtHead(value interface{}) *List {
	l.AddAt(value, &l.root)

	return l
}

func (l *List) Len() int {
	return l.len
}

func (l *List) Reverse() *List {
	current := l.root.next
	prev := &Element{}
	for i := 0; i < l.Len(); i++ {
		next := current.next
		current.next = prev
		prev = current
		current = next

	}

	l.root.next = prev

	return l
}

func TestReverse(test *testing.T) {
	l := new(List).Init()
	l = l.AddAtHead(0)
	l = l.AddAtHead(4)
	l = l.AddAtHead(3)
	l = l.AddAtHead(1)

	l = l.Reverse()

	head := l.root.next
	for i := 0; i < l.Len(); i++ {
		fmt.Println(head.Value) // 0 4 3 1
		head = head.next
	}
}

func TestAddAtHead(test *testing.T) {
	l := new(List).Init()
	l = l.AddAtHead(0)
	l = l.AddAtHead(4)
	l = l.AddAtHead(3)
	l = l.AddAtHead(1)

	head := l.root.next
	for i := 0; i < l.Len(); i++ {
		fmt.Println(head.Value) // 1 3 4 0
		head = head.next
	}
}

type PQueueItem struct {
	value int
	index int
}
type PQueue []*PQueueItem

func (pq PQueue) Len() int {
	return len(pq)
}

func (pq PQueue) Less(i, j int) bool {
	return pq[i].value < pq[j].value
}

func (pq PQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push会修改队列数据，使用指针作为函数receiver
func (pq *PQueue) Push(x interface{}) {
	i := x.(*PQueueItem)
	i.index = len(*pq)
	*pq = append(*pq, i)
}

func (pq *PQueue) Pop() interface{} {
	n := len(*pq)
	item := (*pq)[n-1]
	item.index = -1 // for safety
	*pq = (*pq)[0:(n - 1)]

	return item
}

type Interface2 interface {
	sort.Interface

	Pop() interface{}
}

func Init2(data Interface2) {
	n := data.Len()
	// 如果一颗二叉树最多只有最下面的两层结点度数可以小于2，并且最下面一层的结点都集中在该层最左边的连续位置上，
	// 则此二叉树称做完全二叉树（complete binary tree）
	// 即在有n个结点的完全二叉树中，当 i>n/2-1 时，以i结点为根的子树已经是堆
	// 还必须是n/2-1开始到0，写n或者n/2都行
	for i := n / 2; i >= 0; i-- {
		down2(i, data)
	}
}

func down2(index int, data Interface2) {
	for {
		left := 2*index + 1
		if left >= data.Len() {
			break
		}

		child := left
		right := left + 1
		if right < data.Len() && data.Less(right, left) {
			child = right
		}

		if data.Less(index, child) {
			break
		}

		data.Swap(index, child)
		index = child
	}
}

// 待删除元素与最后元素交换，然后删除
func Pop2(data Interface2) interface{} {
	i := 0
	tail := data.Len() - 1
	data.Swap(i, tail)
	value := data.Pop()

	down2(i, data)

	return value
}

func mergeKLists(lists []*List) *List {
	var pq PQueue
	tmp := 0
	for _, list := range lists {
		head := list.root.next
		for i := tmp; i < list.Len()+tmp; i++ {
			pq = append(pq, &PQueueItem{
				index: i,
				value: head.Value.(int),
			})
			head = head.next
		}

		tmp = list.Len()
	}

	//heap.Init(&pq)
	Init2(&pq)
	for pq.Len() > 0 {
		//item := heap.Pop(&pq).(*PQueueItem)
		item := Pop2(&pq).(*PQueueItem)
		fmt.Println(item.value)
	}

	return nil
}

func TestMergeKLists(test *testing.T) {
	l1 := new(List).Init()
	l1 = l1.AddAtHead(0)
	l1 = l1.AddAtHead(5)
	l1 = l1.AddAtHead(4)
	l1 = l1.AddAtHead(1)

	l2 := new(List).Init()
	l2 = l2.AddAtHead(4)
	l2 = l2.AddAtHead(3)
	l2 = l2.AddAtHead(1)

	l3 := new(List).Init()
	l3 = l3.AddAtHead(6)
	l3 = l3.AddAtHead(2)

	lists := []*List{l1, l2, l3}
	list := mergeKLists(lists)

	head := list.root.next
	for i := 0; i < list.Len(); i++ {
		fmt.Println(head.Value)
		head = head.next
	}
}
