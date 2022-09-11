package leetcode

import (
	"gotest.tools/assert"
	"math"
	"testing"
)

func reverseInt(x int) int {
	negative := false
	var cur int
	if x < 0 {
		negative = true
		cur = x * -1
	} else {
		cur = x
	}

	var offsets []int
	for cur != 0 {
		offsets = append(offsets, cur%10)
		cur = cur / 10
	}

	result := 0
	for _, offset := range offsets {
		if offset == 0 && result == 0 {
			continue
		}

		result = result*10 + offset
	}

	if negative {
		result = result * -1
	}

	if float64(result) < (math.Pow(2, 31)*-1) || float64(result) > (math.Pow(2, 31)-1) {
		return 0
	}

	return result
}

func TestReverseInt(test *testing.T) {
	assert.Equal(test, 321, reverseInt(123))
	assert.Equal(test, -321, reverseInt(-123))
	assert.Equal(test, 21, reverseInt(120))
	assert.Equal(test, 0, reverseInt(0))
	assert.Equal(test, 0, reverseInt(1534236469))
}
