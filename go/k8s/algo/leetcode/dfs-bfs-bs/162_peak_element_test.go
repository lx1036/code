package dfs_bfs_bs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/find-peak-element/solution/

func findPeakElement(nums []int) int {
	return bs(nums)
}

func peak(nums []int) int {
	for i := 0; i < len(nums); i++ {
		if nums[i] > nums[i+1] {
			return i
		}
	}

	return len(nums) - 1
}

func bs(nums []int) int {
	return bsRecursive(nums, 0, len(nums)-1)
}

// 二分查找模板：https://leetcode-cn.com/leetbook/read/binary-search/xerqxt/
func bsRecursive(nums []int, left, right int) int {
	if left == right {
		return left
	}

	mid := (left + right) / 2
	if nums[mid] > nums[mid+1] {
		return bsRecursive(nums, left, mid)
	}

	return bsRecursive(nums, mid+1, right)
}

func TestPeakElement(test *testing.T) {
	nums := []int{1, 2, 3, 1}
	ans := findPeakElement(nums)
	assert.Equal(test, 2, ans)

	nums = []int{1, 2, 1, 3, 5, 6, 4}
	ans = findPeakElement(nums)
	assert.Equal(test, 5, ans)

	nums = []int{4, 3, 2, 1, 2, 3, 4}
	ans = findPeakElement(nums)
	assert.Equal(test, 6, ans)
}
