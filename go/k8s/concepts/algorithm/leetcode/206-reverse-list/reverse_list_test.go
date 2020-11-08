package main

import (
	"fmt"
	"testing"
)

type ListNode struct {
	value int
	next *ListNode
}

// 迭代
// 时间复杂度O(n)，空间复杂度O(1)
func ReverseIteration(l1 *ListNode) *ListNode {
	current := l1
	var prev *ListNode = nil
	for current != nil {
		next := current.next
		current.next = prev
		prev = current
		current = next
	}

	return prev
}

// 递归
func ReverseRecursion(l1 *ListNode) *ListNode {
	current := l1

}

func TestReverse(test *testing.T) {
	a1 := &ListNode{value: 3, next: nil}
	a2 := &ListNode{value: 4, next: a1}
	a3 := &ListNode{value: 2, next: a2}

	c := ReverseIteration(a3)
	for c != nil {
		fmt.Println(c.value)
		c = c.next
	}
}
