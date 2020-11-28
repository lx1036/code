package dfs_bfs_bs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/search-in-rotated-sorted-array/

// 旋转数组
// 时间复杂度O(logN)，空间复杂度O(1)
func search704(nums []int, target int) int {
	left, right := 0, len(nums)-1
	ans := -1
	for left <= right {
		mid := left + (right-left)>>1
		if nums[mid] == target {
			ans = mid
			break
		}

		if nums[mid] >= nums[left] {
			if target < nums[mid] && target >= nums[left] {
				right = mid - 1
			} else {
				left = mid + 1
			}
		} else if nums[mid] <= nums[right] {
			if target > nums[mid] && target <= nums[right] {
				left = mid + 1
			} else {
				right = mid - 1
			}
		}
	}

	return ans
}

func TestSearch704(test *testing.T) {
	nums := []int{4, 5, 6, 7, 0, 1, 2}
	target := 0
	ans := search704(nums, target)
	assert.Equal(test, 4, ans)

	nums = []int{1, 3, 5}
	target = 2
	ans = search704(nums, target)
	assert.Equal(test, -1, ans)
}
