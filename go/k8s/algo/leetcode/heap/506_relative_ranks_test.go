package heap

import (
	"fmt"
	"strconv"
	"testing"

	"k8s.io/klog/v2"
)

// https://leetcode.cn/problems/relative-ranks/solution/zui-da-dui-by-lx1036-gc9u/

func findRelativeRanks(score []int) []string {
	pq := NewPriorityQueue506()
	for key, value := range score {
		pq.Push(&Item{
			key:   key,
			value: value,
		})
	}
	result := map[int]string{}
	i := 0
	for len(pq.items) != 0 {
		item := pq.Pop()
		var value string
		if i == 0 {
			value = "Gold Medal"
		} else if i == 1 {
			value = "Silver Medal"
		} else if i == 2 {
			value = "Bronze Medal"
		} else {
			value = strconv.Itoa(i + 1)
		}

		i++
		result[item.key] = value
	}
	var results []string
	for key, _ := range score {
		results = append(results, result[key])
	}

	return results
}

func TestFindRelativeRanks(test *testing.T) {
	score := []int{5, 4, 3, 2, 1}
	result := findRelativeRanks(score)
	klog.Info(result) // ["Gold Medal","Silver Medal","Bronze Medal","4","5"]

	score = []int{10, 3, 8, 9, 4}
	result = findRelativeRanks(score) // ["Gold Medal","5","Bronze Medal","Silver Medal","4"]
	klog.Info(result)
}

type Item struct {
	key   int
	value int
}

func (item *Item) Less(than *Item) bool {
	return item.value > than.value // 最大堆
	//return item.value <= than.value // 最小堆
}

type PriorityQueue506 struct {
	items []*Item
}

func NewPriorityQueue506() *PriorityQueue506 {
	return &PriorityQueue506{
		items: make([]*Item, 0),
	}
}

func (pq *PriorityQueue506) Push(item *Item) {
	pq.items = append(pq.items, item)
	pq.up(len(pq.items) - 1)
}

func (pq *PriorityQueue506) Pop() *Item {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueue506) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueue506) down(index int) {
	for {
		left := 2*index + 1
		l := len(pq.items) - 1
		if left > l {
			break
		}
		right := left + 1
		min := left
		if right <= l && pq.items[right].Less(pq.items[min]) {
			min = right
		}

		if pq.items[index].Less(pq.items[min]) {
			break
		}

		pq.swap(index, min)
		index = min
	}
}

func (pq *PriorityQueue506) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}

func TestPriorityQueue506(test *testing.T) {
	pq := NewPriorityQueue506()
	data := []*Item{ // 完全二叉树：3 6 15 10 7 20 30 17 19
		&Item{value: 3},
		&Item{value: 10},
		&Item{value: 15},
		&Item{value: 20},
		&Item{value: 30},
		&Item{value: 19},
		&Item{value: 17},
		&Item{value: 6},
		&Item{value: 7},
	}
	for _, value := range data { // 3 10 15 20 30 19 17 6 7
		pq.Push(value)
	}
	for len(pq.items) != 0 {
		item := pq.Pop()
		klog.Infof(fmt.Sprintf("%d", item.value)) // 3 6 7 10 15 17 19 20 30
	}
}
