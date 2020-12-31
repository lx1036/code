package linked_list

import (
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/remove-linked-list-elements/

func removeElements(head *ListNode, val int) *ListNode {
	p1 := head

	if p1 == nil {
		return p1
	}
	if p1.Next == nil {
		if p1.Val == val {
			return nil
		} else {
			return p1
		}
	}

	for p1 != nil {
		if p1.Val == val {
			p1 = p1.Next
			head = p1
		}

		next := p1.Next
		if next != nil && next.Val == val {
			p1.Next = next.Next
		}

		p1 = p1.Next
	}

	if head.Next == nil && head.Val == val {
		return nil
	}

	return head
}

func TestRemoveElements(test *testing.T) {
	n1 := &ListNode{
		Val: 6,
	}
	n2 := &ListNode{
		Val:  5,
		Next: n1,
	}
	n3 := &ListNode{
		Val:  4,
		Next: n2,
	}
	n4 := &ListNode{
		Val:  3,
		Next: n3,
	}
	n5 := &ListNode{
		Val:  6,
		Next: n4,
	}
	n6 := &ListNode{
		Val:  2,
		Next: n5,
	}
	n7 := &ListNode{
		Val:  1,
		Next: n6,
	}
	l1 := removeElements(n7, 6)
	for l1 != nil {
		fmt.Println(l1.Val)

		l1 = l1.Next
	}

	m1 := &ListNode{
		Val: 1,
	}
	m2 := &ListNode{
		Val:  1,
		Next: m1,
	}
	m3 := &ListNode{
		Val:  1,
		Next: m2,
	}
	l2 := removeElements(m3, 1)
	for l2 != nil {
		fmt.Println(l2.Val)

		l2 = l2.Next
	}
}
