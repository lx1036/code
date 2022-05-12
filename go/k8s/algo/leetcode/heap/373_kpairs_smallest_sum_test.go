package heap

import (
	"testing"

	"k8s.io/klog/v2"
)

// https://leetcode-cn.com/problems/find-k-pairs-with-smallest-sums/

func kSmallestPairs(nums1 []int, nums2 []int, k int) [][]int {
	pq := NewPriorityQueue373()

	l := min(len(nums1), len(nums2))
	var mins, maxs []int
	if nums1[0] < nums2[0] {
		mins = nums1
		maxs = nums2
	} else {
		mins = nums2
		maxs = nums1
	}

	for i := 0; i < l; i++ {

		j := i
		for maxs[j] < mins[i] {
			pq.Push(&Item373{
				keys:  []int{mins[i], maxs[j]},
				value: mins[i] + maxs[j],
			})
			j++
		}

	}

	i := 1
	var results [][]int
	for len(pq.items) != 0 {
		item := pq.Pop()
		results = append(results, item.keys)
		if i == k {
			break
		}
		i++
	}

	return results
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestKSmallestPairs(test *testing.T) {
	nums1 := []int{1, 7, 11}
	nums2 := []int{2, 4, 6}
	k := 3
	klog.Info(kSmallestPairs(nums1, nums2, k))

	nums1 = []int{1, 1, 2}
	nums2 = []int{1, 2, 3}
	k = 2
	klog.Info(kSmallestPairs(nums1, nums2, k))

	nums1 = []int{1, 2}
	nums2 = []int{3}
	k = 3
	klog.Info(kSmallestPairs(nums1, nums2, k))
}

type Item373 struct {
	keys  []int
	value int
}

func (item *Item373) Less(than *Item373) bool {
	//return item.value > than.value // 最大堆
	return item.value < than.value // 最小堆
}

// 使用最小堆
type PriorityQueue373 struct {
	items []*Item373
}

func NewPriorityQueue373() *PriorityQueue373 {
	return &PriorityQueue373{
		items: make([]*Item373, 0),
	}
}

func (pq *PriorityQueue373) Push(item *Item373) {
	pq.items = append(pq.items, item)
	pq.up(len(pq.items) - 1)
}

func (pq *PriorityQueue373) Pop() *Item373 {
	result := pq.items[0]
	l := len(pq.items) - 1
	pq.items[0] = pq.items[l]
	pq.items = pq.items[:l]
	pq.down(0)
	return result
}

func (pq *PriorityQueue373) Peek() *Item373 {
	return pq.items[0]
}

func (pq *PriorityQueue373) up(index int) {
	for {
		parent := (index - 1) / 2 // 0==0
		if parent == index || !pq.items[index].Less(pq.items[parent]) {
			break
		}

		pq.swap(index, parent)
		index = parent
	}
}

func (pq *PriorityQueue373) down(index int) {
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

func (pq *PriorityQueue373) swap(index, parent int) {
	pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
}
