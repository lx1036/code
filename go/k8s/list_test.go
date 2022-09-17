package k8s

import (
	"fmt"
	"testing"
)

// (1) 反转单链表
type List2 struct {
	value int
	next  *List2
}

func (l *List2) reverse() *List2 {
	cur := l
	var prev *List2
	for cur != nil {
		next := cur.next
		cur.next = prev
		prev = cur
		cur = next
	}
	return prev
}

func findDuplicate(data []int) int {
	tmp := make(map[int]int)
	l := len(data)
	for i := 0; i < l; i++ {
		tmp[data[i]] = tmp[data[i]] + 1
		if tmp[data[i]] == 2 {
			return data[i]
		}
	}

	return -1
}

func TestFindDuplicate(test *testing.T) {
	t1 := []int{2, 3, 1, 0, 2, 5, 3}
	fmt.Println(findDuplicate(t1))
}
