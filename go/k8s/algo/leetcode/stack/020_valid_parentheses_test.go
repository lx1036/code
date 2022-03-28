package stack

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// https://leetcode-cn.com/problems/valid-parentheses/

// 使用栈
func isValid(s string) bool {
	l := len(s)
	if l%2 != 0 {
		return false
	}

	var stack []byte
	for i := 0; i < l; i++ {
		if s[i] == '(' {
			stack = append(stack, ')')
		} else if s[i] == '[' {
			stack = append(stack, ']')
		} else if s[i] == '{' {
			stack = append(stack, '}')
		} else if len(stack) == 0 {
			return false
		} else {
			if s[i] != stack[len(stack)-1] {
				return false
			} else {
				stack = stack[:(len(stack) - 1)]
			}
		}
	}

	if len(stack) > 0 {
		return false
	}

	return true
}

func TestValid(test *testing.T) {
	s := "()"
	assert.Equal(test, true, isValid(s))

	s = "()[]{}"
	assert.Equal(test, true, isValid(s))

	s = "(]"
	assert.Equal(test, false, isValid(s))

	s = "([)]"
	assert.Equal(test, false, isValid(s))

	s = "{[]}"
	assert.Equal(test, true, isValid(s))

	s = "(("
	assert.Equal(test, false, isValid(s))
}

func valid(queue []int) bool {
	if len(queue)%2 != 0 {
		return false
	}

	l := len(queue)
	for i, j := 0, l-1; j >= i; i, j = i+1, j-1 {
		if queue[i] != queue[j] {
			return false
		}
	}
	return true
}
