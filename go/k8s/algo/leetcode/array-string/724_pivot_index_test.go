package array_string

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/find-pivot-index/

func pivotIndex(nums []int) int {
	l := len(nums)

	sum := 0
	for i := 0; i < l; i++ {
		sum += nums[i]
	}

	suml := 0
	for i := 0; i < l; i++ {
		if (sum - nums[i] - suml) == suml { // 总和减去左边和，减去当前值等于左边后，该值就是中心索引
			return i
		}

		suml += nums[i]
	}

	return -1
}

func TestPivotIndex(test *testing.T) {
	nums := []int{1, 7, 3, 6, 5, 6}
	result := pivotIndex(nums)
	assert.Equal(test, 3, result)
}
