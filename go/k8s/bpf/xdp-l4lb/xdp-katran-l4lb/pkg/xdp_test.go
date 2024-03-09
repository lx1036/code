package pkg

import (
	"fmt"
	"testing"
)

// 主要用来验证一些简单的逻辑

type Person struct {
	nums []int
}

func (p *Person) hash(nums []int) {
	for i := 0; i < len(nums); i++ {
		p.nums[i] = nums[i]
	}
}

func TestName(test *testing.T) {
	person := Person{
		nums: []int{1, 2, 3},
	}

	person.hash([]int{4, 5, 6})

	fmt.Println(person.nums)
}
