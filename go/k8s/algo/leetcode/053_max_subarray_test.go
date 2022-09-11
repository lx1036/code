package leetcode

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func maxSubArray(nums []int) int {
	result := 0
	l := len(nums)
	if l == 1 {
		return nums[0]
	}

	for i := 0; i < l; i++ {
		tmp := nums[i]
		for j := i + 1; j < l; j++ {
			tmp = tmp + nums[j]
			if tmp > result {
				result = tmp
			}
		}
	}

	return result
}

func TestMaxSubArray(test *testing.T) {
	nums := []int{-2, 1, -3, 4, -1, 2, 1, -5, 4}
	assert.Equal(test, 6, maxSubArray(nums))
	nums = []int{1}
	assert.Equal(test, 1, maxSubArray(nums))
	nums = []int{5, 4, -1, 7, 8}
	assert.Equal(test, 23, maxSubArray(nums))

}
