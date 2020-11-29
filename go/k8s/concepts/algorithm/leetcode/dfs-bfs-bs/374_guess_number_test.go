package dfs_bfs_bs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/guess-number-higher-or-lower/

var PICK = 6

func guess(num int) int {
	pick := PICK
	if pick < num {
		return -1
	} else if pick > num {
		return 1
	} else {
		return 0
	}
}

func guessNumber(n int) int {
	low, right := 0, n
	ans := -1
	for low <= right {
		mid := low + (right-low)>>1
		tmp := guess(mid)
		if tmp == 1 || tmp == 0 {
			ans = mid
			low = mid + 1
		} else if tmp == -1 {
			right = mid - 1
		}
	}

	return ans
}

func TestGuessNumber(test *testing.T) {
	ans := guessNumber(10)
	assert.Equal(test, 6, ans)
}
