package dfs_bfs_bs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/first-bad-version/

func firstBadVersion(n int) int {
	left := 1
	right := n
	for left < right {
		mid := left + (right-left)/2
		if isBadVersion(mid) {
			right = mid
		} else {
			left = mid + 1
		}
	}

	return left
}

func isBadVersion(version int) bool {
	return version >= 4
}

func TestFirstBadVersion(test *testing.T) {
	ans := firstBadVersion(5)
	assert.Equal(test, 4, ans)
}
