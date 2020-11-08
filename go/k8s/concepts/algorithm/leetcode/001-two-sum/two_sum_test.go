package main

import (
	"fmt"
	"testing"
)

// 使用哈希表来O(1)查找target-value的值
// 总体复杂度：
// 时间复杂度是O(n)，对于每一个元素x，可以O(1)查找target-x值
// 空间复杂度O(n)，主要是哈希表的开销
func TwoSum(nums []int, target int) []int {
	hashTable := map[int]int{}

	for key, value := range nums {
		if p, ok := hashTable[target - value];ok {
			return []int{key, p}
		}

		hashTable[value] = key
	}

	return nil
}

func TwoSum2(nums []int, target int) []int  {
	hashTable := map[int]int{}
	for key, value := range nums {
		hashTable[value] = key
	}

	for key, value := range nums {
		if p, ok := hashTable[target - value]; ok {
			return []int{p, key}
		}
	}

	return nil
}


func TestTwoSum(test *testing.T) {
	nums := []int{6, 3, 8, 2, 1}
	target := 8
	result := TwoSum(nums, target)
	fmt.Println(result)

	result2 := TwoSum2(nums, target)
	fmt.Println(result2)
}

