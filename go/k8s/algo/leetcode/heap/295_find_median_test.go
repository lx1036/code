package heap

import (
	"container/heap"
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/find-median-from-data-stream/

// 字节跳动面试题：数据流的中位数，two-heaps解决

type PriorityQueueMin []int

func (min PriorityQueueMin) Len() int {
	return len(min)
}

func (min PriorityQueueMin) Less(i, j int) bool {
	return min[i] < min[j]
}

func (min PriorityQueueMin) Swap(i, j int) {
	min[i], min[j] = min[j], min[i]
}

func (min *PriorityQueueMin) Push(x interface{}) {
	*min = append(*min, x.(int))
}

func (min *PriorityQueueMin) Pop() interface{} {
	old := *min
	n := len(old)
	x := old[n-1]
	*min = old[0 : n-1]

	return x
}

type PriorityQueueMax []int

func (max PriorityQueueMax) Len() int {
	return len(max)
}

func (max PriorityQueueMax) Less(i, j int) bool {
	return max[i] > max[j]
}

func (max PriorityQueueMax) Swap(i, j int) {
	max[i], max[j] = max[j], max[i]
}

func (max *PriorityQueueMax) Push(x interface{}) {
	*max = append(*max, x.(int))
}

func (max *PriorityQueueMax) Pop() interface{} {
	old := *max
	n := len(old)
	x := old[n-1]
	*max = old[0 : n-1]

	return x
}

type MedianFinder struct {
	min *PriorityQueueMin // 小顶堆，都是较大的数

	max *PriorityQueueMax // 大顶堆，都是较小的数
}

/** initialize your data structure here. */
func Constructor() *MedianFinder {
	finder := &MedianFinder{
		min: &PriorityQueueMin{},
		max: &PriorityQueueMax{},
	}

	heap.Init(finder.min)
	heap.Init(finder.max)

	return finder
}

// 必须保证：大顶堆len - 小顶堆len == 1 or 0
func (this *MedianFinder) AddNum(num int) {
	heap.Push(this.max, num)
	heap.Push(this.min, heap.Pop(this.max)) // 调整两个堆平衡，此时从大顶堆Pop出最大元素，加入到小顶堆
	for this.max.Len() < this.min.Len() {   // 平衡调整，|小顶堆len-大顶堆len|==1
		heap.Push(this.max, heap.Pop(this.min))
	}
}

func (this *MedianFinder) FindMedian() float64 {
	fmt.Println(this.min.Len(), this.max.Len())
	if this.max.Len() > this.min.Len() { // 大顶堆len-小顶堆len==1，中位数直接就是大顶堆top
		return float64(heap.Pop(this.max).(int))
	}

	// 大顶堆len==小顶堆len，两个top和 * 0.5
	return float64(heap.Pop(this.min).(int)+heap.Pop(this.max).(int)) * 0.5
}

func TestFindMedian(test *testing.T) {
	median := Constructor()
	median.AddNum(2)
	median.AddNum(1)
	median.AddNum(5)
	median.AddNum(4)
	median.AddNum(3)

	fmt.Println(median.FindMedian())

	median1 := Constructor()
	median1.AddNum(1)
	median1.AddNum(2)
	median1.AddNum(3)
	fmt.Println(median1.FindMedian())
}
