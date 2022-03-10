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
