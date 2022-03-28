package dfs_bfs_bs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/find-minimum-in-rotated-sorted-array/

func findMin(nums []int) int {
	if len(nums) == 1 {
		return nums[0]
	}

	left, right := 0, len(nums)-1
	if nums[right] > nums[left] {
		return nums[left]
	}
	mid := 0
	for left <= right {
		mid = (left + right) / 2

		if nums[mid] > nums[mid+1] {
			return nums[mid+1]
		}
		if nums[mid-1] > nums[mid] {
			return nums[mid]
		}

		if nums[mid] > nums[0] {
			left = mid + 1
		} else if nums[mid] <= nums[0] {
			right = mid - 1
		}
	}

	return -1
}

func TestFindMin(test *testing.T) {
	nums := []int{3, 4, 5, 1, 2}
	ans := findMin(nums)
	assert.Equal(test, 1, ans)

	nums = []int{4, 5, 6, 7, 0, 1, 2}
	ans = findMin(nums)
	assert.Equal(test, 0, ans)

	nums = []int{1}
	ans = findMin(nums)
	assert.Equal(test, 1, ans)

	nums = []int{3, 1, 2}
	ans = findMin(nums)
	assert.Equal(test, 1, ans)

	nums = []int{11, 13, 15, 17}
	ans = findMin(nums)
	assert.Equal(test, 11, ans)
}
