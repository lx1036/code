package heap

import "sort"

// **[堆 堆排序 优先队列 图文详解](https://www.cnblogs.com/yahuian/p/11945144.html)**

type Interface interface {
	sort.Interface

	Push(x interface{})
	Pop() interface{}
}

func Init(data Interface) {
	n := data.Len()
	for i := n/2 - 1; i >= 0; i-- {
		down(data, i, n)
	}
}

func Push(data Interface, x interface{}) {
	data.Push(x)
	up(data, data.Len()-1)
}

// 删除，首先把最末端的结点与要删除节点的交换位置
func Pop(data Interface) interface{} {

	// TODO: 为何要交换首位元素后，在下沉，最后才pop，感觉交换后下沉不是变回原来的tree么
	// 不应该是pop后
	n := data.Len() - 1
	data.Swap(0, n) // 0索引是要删除的
	down(data, 0, n)

	return data.Pop()
}

// 删除堆中位置为i的元素
func Remove(data Interface, i int) {
	n := data.Len() - 1
	data.Swap(i, n) // 用最后元素和待删除元素交换

}

func Fix(data Interface, i int) {

}

// 自下而上比较和交换新节点和父节点位置，直到满足最小堆性质，即子节点小于等于父节点
func up(data Interface, child int) {
	for {
		// slice里的i位置，左子节点为2i+1，右子节点2i+2
		parent := (child - 1) / 2                        // 父节点位置
		if child == parent || data.Less(parent, child) { // 要么升到最小堆顶端，要么父节点值比新插入的节点小
			break
		}
		data.Swap(parent, child) // 交换父节点和子节点位置
		child = parent
	}
}

// 自上而下比较和交换
// n是数组元素最后一位索引 len(data)-1, a是当前元素索引
func down(data Interface, a, n int) {
	current := a // 父节点索引
	for {
		left := 2*current + 1 // 左子节点
		if left >= n {
			break // current已经是叶子节点
		}

		// 求解child=min(left, right)
		child := left
		if right := left + 1; right < n && data.Less(right, left) {
			child = right
		}

		if data.Less(current, child) {
			break // 当前节点小于 min(左子节点, 右子节点)
		}

		data.Swap(current, child)
		current = child
	}
}

// /usr/local/go/src/sort/sort.go
// 三种排序算法

func heapSort(data Interface, a, b int) {

}

func insertionSort(data Interface, a, b int) {

}

func quickSort(data Interface, a, b, maxDepth int) {

}
