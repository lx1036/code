package leetcode

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/klog/v2"
	"sort"
	"testing"
)

// https://leetcode-cn.com/problems/merge-sorted-array/

// merge([]int{4, 5, 6, 10, 11, 12, 13, 27}, []int{1, 3, 8, 9, 25, 30})
// -> []int{1, 3, 4, 5, 6, 8, 9, 10, 11, 12, 13, 25, 27, 30}

func merge(nums1 []int, nums2 []int) []int {
	var result []int
	if len(nums1) == 0 {
		copy(result, nums2)
		return result
	}
	if len(nums2) == 0 {
		copy(result, nums1)
		return result
	}

	leader, follower := nums1, nums2
	if nums2[0] < nums1[0] {
		leader, follower = nums2, nums1
	}

	for len(leader) > 0 {
		n := sort.Search(len(leader), func(i int) bool {
			return leader[i] > follower[0]
		})
		result = append(result, leader[:n]...)
		if n >= len(leader) {
			break
		}

		leader, follower = follower, leader[n:]
	}

	result = append(result, follower...)

	return result
}

func TestMerge(test *testing.T) {
	nums1 := []int{4, 5, 6, 10, 11, 12, 13, 27}
	nums2 := []int{1, 3, 8, 9, 25, 30}
	result := []int{1, 3, 4, 5, 6, 8, 9, 10, 11, 12, 13, 25, 27, 30}
	r := merge(nums1, nums2)
	if !assert.Equal(test, result, r) {
		klog.Fatal("r not equal")
	}

	r2 := append(nums1, nums2...)
	sort.Slice(r2, func(i, j int) bool {
		return r2[i] < r2[j]
	})
	if !assert.Equal(test, result, r2) {
		klog.Fatal("r2 not equal")
	}
}
