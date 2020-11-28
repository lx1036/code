package dfs_bfs_bs

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/search-insert-position/solution/sou-suo-cha-ru-wei-zhi-by-leetcode-solution/

// 时间复杂度O(logN)，空间复杂度O(1)，我们只需要常数空间存放若干变量
func searchInsert(nums []int, target int) int {
	n := len(nums)
	left, right := 0, n-1
	ans := n
	for left <= right {
		mid := (right-left)>>1 + left
		if target <= nums[mid] {
			ans = mid
			right = mid - 1
		} else {
			left = mid + 1
		}
	}

	return ans
}

func TestSearchInsert(test *testing.T) {
	right := 10
	left := 3
	mid := (right - left) >> 1 // 除2取整 (right - left)/2
	fmt.Println(mid)

	nums := []int{1, 3, 5, 6}
	ans := searchInsert(nums, 5)
	assert.Equal(test, 2, ans)

	ans = searchInsert(nums, 2)
	assert.Equal(test, 1, ans)

	ans = searchInsert(nums, 0)
	assert.Equal(test, 0, ans)
}
