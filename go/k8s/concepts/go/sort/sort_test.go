package sort

import (
	"fmt"
	"sort"
	"testing"
)

func TestSearch(test *testing.T) {
	data := []int{1, 3, 5, 7, 9}
	i := sort.Search(len(data), func(i int) bool {
		return data[i] > 12 // output: 5, 如果都比 12 小，就返回 len(data)，表示 would be inserted
	})

	i2 := sort.Search(len(data), func(i int) bool {
		return data[i] > 5 // output: 3
	})

	fmt.Println(i, i2)

	idx := 2
	number := copy(data, data[idx:])
	fmt.Println(number, data) // 3 [5 7 9 7 9]

}
