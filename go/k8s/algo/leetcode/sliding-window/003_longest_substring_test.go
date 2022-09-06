package sliding_window

import (
	"fmt"
	"gotest.tools/assert"
	"testing"
)

// https://leetcode-cn.com/problems/longest-substring-without-repeating-characters/

func lengthOfLongestSubstring(s string) int {
	l := len(s)
	hashTable := map[byte]bool{}
	result := 0

	j := 0
	for i := 0; i < l; i++ {
		for j < l {
			if _, ok := hashTable[s[j]]; !ok {
				hashTable[s[j]] = true
				j++

				if len(hashTable) > result {
					result = len(hashTable)
				}
			} else {
				break
			}
		}

		delete(hashTable, s[i])
	}

	return result
}

// 无重复最长子串用这个函数，滑动窗口方法，两个指针解决
func lengthOfLongestSubstring2(s string) int {
	result := 0
	storage := make(map[byte]int)
	l := len(s)
	for i := 0; i < l; i++ {
		tmp := 0
		for j := i; j < l; j++ {
			if _, ok := storage[s[j]]; !ok {
				storage[s[j]] = j
				tmp++
			} else {
				break
			}
		}

		if tmp > result {
			result = tmp
		}

		storage = map[byte]int{}
	}

	return result
}

func TestLengthOfLongestSubstring2(test *testing.T) {
	hashTable := map[byte]int{}
	fmt.Println(hashTable['a']) // 0

	s := "abcabcbb"
	assert.Equal(test, 3, lengthOfLongestSubstring2(s))

	s = "bbbbb"
	assert.Equal(test, 1, lengthOfLongestSubstring2(s))

	s = "pwwkew"
	assert.Equal(test, 3, lengthOfLongestSubstring2(s))

	s = "ab"
	assert.Equal(test, 2, lengthOfLongestSubstring2(s))

	s = " "
	assert.Equal(test, 1, lengthOfLongestSubstring2(s))

	s = "dvdf"
	assert.Equal(test, 3, lengthOfLongestSubstring2(s))
}

func TestLengthOfLongestSubstring(test *testing.T) {
	hashTable := map[byte]int{}
	fmt.Println(hashTable['a']) // 0

	s := "abcabcbb"
	assert.Equal(test, 3, lengthOfLongestSubstring(s))

	s = "bbbbb"
	assert.Equal(test, 1, lengthOfLongestSubstring(s))

	s = "pwwkew"
	assert.Equal(test, 3, lengthOfLongestSubstring(s))

	s = "ab"
	assert.Equal(test, 2, lengthOfLongestSubstring(s))

	s = " "
	assert.Equal(test, 1, lengthOfLongestSubstring(s))

	s = "dvdf"
	assert.Equal(test, 3, lengthOfLongestSubstring(s))
}
