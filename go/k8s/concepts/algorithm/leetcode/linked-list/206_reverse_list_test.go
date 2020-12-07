package linked_list

import (
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/reverse-linked-list/
//type ListNode struct {
//	value int
//	next  *ListNode
//}

// 迭代
// 时间复杂度O(n)，空间复杂度O(1)
func ReverseIteration(l1 *ListNode) *ListNode {
	current := l1
	var prev *ListNode = nil
	for current != nil {
		next := current.Next
		current.Next = prev
		prev = current
		current = next
	}

	return prev
}

// 递归
func ReverseRecursion(l1 *ListNode) *ListNode {

	return nil
}

func TestReverse(test *testing.T) {
	a1 := &ListNode{Val: 3, Next: nil}
	a2 := &ListNode{Val: 4, Next: a1}
	a3 := &ListNode{Val: 2, Next: a2}

	c := ReverseIteration(a3)
	for c != nil {
		fmt.Println(c.Val)
		c = c.Next
	}
}

func reverseList(root *ListNode) *ListNode {
	current := root
	prev := &ListNode{}

	for current != nil {
		next := current.Next
		current.Next = prev
		prev = current
		current = next
	}

	return prev
}
