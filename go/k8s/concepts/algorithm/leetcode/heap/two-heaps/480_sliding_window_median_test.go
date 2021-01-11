package two_heaps

import (
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/sliding-window-median/

func medianSlidingWindow(nums []int, k int) []float64 {
	i := 0
	j := i + k - 1

	var median []float64
	for slow, fast := i, j; fast < len(nums); slow, fast = slow+1, fast+1 {
		m := Constructor()
		for x := slow; x <= fast; x++ {
			m.AddNum(nums[x])
		}

		median = append(median, m.FindMedian())
	}

	return median
}

func TestMedianSlidingWindow(test *testing.T) {
	nums := []int{1, 3, -1, -3, 5, 3, 6, 7}
	k := 3
	fmt.Println(medianSlidingWindow(nums, k)) // [1 -1 -1 3 5 6]
}
