package dfs_bfs_bs

// https://leetcode-cn.com/problems/find-first-and-last-position-of-element-in-sorted-array/

// 时间复杂度O(logN)
func searchRange(nums []int, target int) []int {
	n := len(nums)
	left, right := 0, n-1
	
	for left < right {
		
		mid := left + (right - left) >> 1
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

