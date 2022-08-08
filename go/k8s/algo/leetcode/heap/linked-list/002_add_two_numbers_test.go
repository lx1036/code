package linked_list

import (
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/add-two-numbers/

type ListNode002 struct {
	Val  int
	Next *ListNode002
}

func AddTwoNumbers(l1 *ListNode002, l2 *ListNode002) *ListNode002 {
	var head, tail *ListNode002
	carry := 0
	for l1 != nil || l2 != nil {
		n1, n2 := 0, 0
		if l1 != nil {
			n1 = l1.Val
			l1 = l1.Next
		}
		if l2 != nil {
			n2 = l2.Val
			l2 = l2.Next
		}

		sum := n1 + n2 + carry
		sum, carry = sum%10, sum/10
		if head == nil {
			head = &ListNode002{Val: sum}
			tail = head
		} else {
			tail.Next = &ListNode002{Val: sum}
			tail = tail.Next
		}
	}

	if carry > 0 {
		tail.Next = &ListNode002{Val: carry}
	}

	return head
}

func TestAddTwoNumbers(test *testing.T) {
	a1 := &ListNode002{Val: 3, Next: nil}
	a2 := &ListNode002{Val: 4, Next: a1}
	a3 := &ListNode002{Val: 2, Next: a2}

	b1 := &ListNode002{Val: 4, Next: nil}
	b2 := &ListNode002{Val: 6, Next: b1}
	b3 := &ListNode002{Val: 5, Next: b2}

	c := AddTwoNumbers(a3, b3)
	for c != nil {
		fmt.Println(c.Val)
		c = c.Next
	}

	A1 := &ListNode002{Val: 3, Next: nil}
	A2 := &ListNode002{Val: 4, Next: A1}
	A3 := &ListNode002{Val: 2, Next: A2}

	B1 := &ListNode002{Val: 6, Next: nil}
	B2 := &ListNode002{Val: 6, Next: B1}
	B3 := &ListNode002{Val: 7, Next: B2}

	C := AddTwoNumbers(A3, B3)
	for C != nil {
		fmt.Println(C.Val)
		C = C.Next
	}
}
