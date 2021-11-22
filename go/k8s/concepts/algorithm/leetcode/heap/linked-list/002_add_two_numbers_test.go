package linked_list

import (
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/add-two-numbers/

func AddTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	var head, tail *ListNode
	carry := 0
	for l1 != nil || l2 != nil {
		n1, n2 := 0, 0

		if l1 != nil {
			n1 = l1.value
			l1 = l1.next
		}
		if l2 != nil {
			n2 = l2.value
			l2 = l2.next
		}

		sum := n1 + n2 + carry
		sum, carry = sum%10, sum/10
		if head == nil {
			head = &ListNode{value: sum}
			tail = head
		} else {
			tail.next = &ListNode{value: sum}
			tail = tail.next
		}
	}

	if carry > 0 {
		tail.next = &ListNode{value: carry}
	}

	return head
}

func TestAddTwoNumbers(test *testing.T) {
	a1 := &ListNode{value: 3, next: nil}
	a2 := &ListNode{value: 4, next: a1}
	a3 := &ListNode{value: 2, next: a2}

	b1 := &ListNode{value: 4, next: nil}
	b2 := &ListNode{value: 6, next: b1}
	b3 := &ListNode{value: 5, next: b2}

	c := AddTwoNumbers(a3, b3)
	for c != nil {
		fmt.Println(c.value)
		c = c.next
	}

	A1 := &ListNode{value: 3, next: nil}
	A2 := &ListNode{value: 4, next: A1}
	A3 := &ListNode{value: 2, next: A2}

	B1 := &ListNode{value: 6, next: nil}
	B2 := &ListNode{value: 6, next: B1}
	B3 := &ListNode{value: 7, next: B2}

	C := AddTwoNumbers(A3, B3)
	for C != nil {
		fmt.Println(C.value)
		C = C.next
	}
}
