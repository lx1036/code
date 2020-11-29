package dfs_bfs_bs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/find-first-and-last-position-of-element-in-sorted-array/

// 时间复杂度O(logN)
func searchRange(nums []int, target int) []int {
	n := len(nums)
	left, right := 0, n-1

	for left < right {
		mid := left + (right-left)>>1
		if target <= nums[mid] {
			if target < nums[mid] {
				right = mid - 1
			} else if target == nums[mid] {
				right = mid
			}
		} else {
			left = mid + 1
		}
	}

	return []int{-1, -1}
}

// https://en.wikipedia.org/wiki/Binary_search_algorithm
// 二分法
func BinarySearch(arr []int, target int) int {
	low, high := 0, len(arr)-1
	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] == target {
			return mid
		} else if target < arr[mid] {
			high = mid - 1
		} else if target > arr[mid] {
			low = mid + 1
		}
	}
	return -1
}

func BinarySearchRecursive(arr []int, low, high, target int) int {
	if low > high {
		return -1
	}
	mid := low + (high-low)/2

	if arr[mid] > target {
		return BinarySearchRecursive(arr, low, mid-1, target)
	} else if arr[mid] < target {
		return BinarySearchRecursive(arr, mid+1, high, target)
	}

	return mid
}

func TestBS(test *testing.T) {
	arr := []int{1, 3, 4, 6, 8, 9}
	target := 6

	ans := BinarySearch(arr, target)
	assert.Equal(test, 3, ans)

	ans = BinarySearchRecursive(arr, 0, len(arr)-1, target)
	assert.Equal(test, 3, ans)
}
