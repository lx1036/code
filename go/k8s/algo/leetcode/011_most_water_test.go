package leetcode

import (
	"gotest.tools/assert"
	"testing"
)

// 应该用滑动窗口双指针方法，类似 003 最长无重复子串

func maxArea2(height []int) int {
	result := 0
	l := len(height)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			w := j - i
			h := min(height[i], height[j])
			s := w * h
			if s > result {
				result = s
			}
		}
	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxArea(height []int) int {
	q := NewPriorityQueue()
	for key, value := range height {
		item := &Item{key: key, value: value}
		q.Push(item)
	}

	l1 := q.Pop()
	l2 := q.Pop()

	width := abs(l2.key - l1.key)
	if width == 0 {
		width = 1
	}

	return width * l2.value
}

func abs(x int) int {
	if x >= 0 {
		return x
	}

	return x * -1
}

type Item struct {
	key, value int
}

func (item *Item) Less(other *Item) bool {
	return item.value > other.value
}

type PriorityQueue struct {
	items []*Item
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
		if index == parent || !q.items[index].Less(q.items[parent]) {
			break
		}
		q.swap(index, parent)
		index = parent
	}
}

func (q *PriorityQueue) down(index int) {
	for {
		left := 2*index + 1
		last := len(q.items) - 1
		if left > last {
			break
		}
		min := left
		right := left + 1
		if right > last || q.items[right].Less(q.items[left]) {
			min = right
		}
		if q.items[index].Less(q.items[min]) {
			break
		}

		q.swap(index, min)
		index = min
	}
}

func (q *PriorityQueue) swap(index, parent int) {
	q.items[index], q.items[parent] = q.items[parent], q.items[index]
}

func TestMaxArea(test *testing.T) {
	input := []int{1, 8, 6, 2, 5, 4, 8, 3, 7}
	assert.Equal(test, 49, maxArea2(input))
	input = []int{1, 1}
	assert.Equal(test, 1, maxArea2(input))
}
