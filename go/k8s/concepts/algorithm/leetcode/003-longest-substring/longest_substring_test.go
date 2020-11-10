package _03_longest_substring

import (
	"fmt"
	"gotest.tools/assert"
	"testing"
)

// https://leetcode-cn.com/problems/longest-substring-without-repeating-characters/

func lengthOfLongestSubstring(s string) int {
	l := len(s)

	hashTable := map[byte]int{}
	result := 0
	
	for i := 0; i < l; i++ {



		if value, ok := hashTable[s[i]]; ok {
			/*if len(hashTable) > result {
				result = len(hashTable)
			}*/
			hashTable = map[byte]int{}
			hashTable[s[i]] = value
		} else {
			hashTable[s[i]]= i
		}

		if len(hashTable) > result {
			result = len(hashTable)
		}
	}

	/*if len(hashTable) > result {
		result = len(hashTable)
	}*/

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


