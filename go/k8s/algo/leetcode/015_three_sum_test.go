package leetcode

import (
	"k8s.io/klog/v2"
	"sort"
	"testing"
)

// https://leetcode.cn/problems/3sum/

func threeSum(nums []int) [][]int {
	// 先排序

	sort.Slice(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})

	var result [][]int
	l := len(nums)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				if nums[i]+nums[j]+nums[k] == 0 {
					result = append(result, []int{nums[i], nums[j], nums[k]})
				}
			}
		}
	}

	return result
}

func TestThreeSum(test *testing.T) {
	nums := []int{-1, 0, 1, 2, -1, -4}
	klog.Info(threeSum(nums)) // {-1,-1,2},{-1,0,1} 顺序不重要，这里值是 [[-1 -1 2] [-1 0 1] [-1 0 1]]
}
