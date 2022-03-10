package dfs_bfs_bs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/sqrtx/

func Sqrt(x int) int {
	ans := 0

	for ans*ans <= x {
		ans += 1
	}

	return ans - 1
}

func SqrtBS(x int) int {
	low, high := 0, x
	ans := -1
	for low <= high {
		mid := low + (high-low)>>1
		if mid*mid > x {
			high = mid - 1
		} else if mid*mid <= x {
			ans = mid
			low = mid + 1
		} else {
			return mid
		}
	}

	return ans
}

func TestSqrt(test *testing.T) {
	ans := Sqrt(10)
	assert.Equal(test, 3, ans)

	ans = Sqrt(4)
	assert.Equal(test, 2, ans)

	ans = Sqrt(8)
	assert.Equal(test, 2, ans)

	ans = SqrtBS(10)
	assert.Equal(test, 3, ans)
	ans = SqrtBS(4)
	assert.Equal(test, 2, ans)
	ans = SqrtBS(8)
	assert.Equal(test, 2, ans)
}
