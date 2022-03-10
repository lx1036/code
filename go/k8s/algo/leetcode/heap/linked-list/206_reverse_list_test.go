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

func TestReverseList(test *testing.T) {
	a1 := &ListNode{Val: 3, Next: nil}
	a2 := &ListNode{Val: 4, Next: a1}
	a3 := &ListNode{Val: 2, Next: a2}

	c := reverseList(a3)
	for c != nil {
		fmt.Println(c.Val)
		c = c.Next
	}
}

func reverseList(head *ListNode) *ListNode {
	if head == nil || head.Next == nil {
		return head
	}

	// head先赋值，下面会改变head.Next指针
	current := head.Next
	// head赋值给prev，修改prev.Next，同时也是修改head.Next
	prev := head
	prev.Next = nil
	for current != nil {
		next := current.Next
		current.Next = prev
		prev = current
		current = next
	}

	return prev

	// 这个也不对，会打印出最后的空节点
	/*current := root
	prev := &ListNode{}

	for current != nil {
		next := current.Next
		current.Next = prev
		prev = current
		current = next
	}

	return prev*/
}
