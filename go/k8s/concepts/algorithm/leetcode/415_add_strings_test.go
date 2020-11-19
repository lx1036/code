package leetcode

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

// i,j两个指针同时移动
func addStrings(num1 string, num2 string) string {
	add := 0
	ans := ""
	for i, j := len(num1)-1, len(num2)-1; i >= 0 || j >= 0 || add != 0; i, j = i-1, j-1 {
		var x, y int // i,j不相等时，x,y补0
		if i >= 0 {
			x = int(num1[i] - '0')
		}
		if j >= 0 {
			y = int(num2[j] - '0')
		}

		result := x + y + add
		ans = strconv.Itoa(result%10) + ans
		add = result / 10
	}

	return ans
}

func TestAddStrings(test *testing.T) {
	result := addStrings("86043", "5582")
	assert.Equal(test, "91625", result)

	result = addStrings("96043", "5582")
	assert.Equal(test, "101625", result)
}
