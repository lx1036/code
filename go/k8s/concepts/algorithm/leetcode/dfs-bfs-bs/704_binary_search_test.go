package dfs_bfs_bs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/binary-search/

// 时间复杂度O(logN)，空间复杂度O(1)
func search(nums []int, target int) int {
	return searchRecursive(nums, 0, len(nums)-1, target)
}

func searchRecursive(nums []int, low, high, target int) int {
	if low > high {
		return -1
	}

	mid := low + (high-low)>>1
	if nums[mid] > target {
		return searchRecursive(nums, low, mid-1, target)
	} else if nums[mid] < target {
		return searchRecursive(nums, mid+1, high, target)
	}

	return mid
}

func TestSearch(test *testing.T) {
	nums := []int{-1, 0, 3, 5, 9, 12}
	target := 9
	ans := search(nums, target)
	assert.Equal(test, 4, ans)

	nums = []int{-1, 0, 3, 5, 9, 12}
	target = 2
	ans = search(nums, target)
	assert.Equal(test, -1, ans)
}
