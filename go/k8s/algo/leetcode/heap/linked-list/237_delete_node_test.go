package linked_list

import (
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/delete-node-in-a-linked-list/

func deleteNode(head, node *ListNode) *ListNode {
	p1 := head

	for p1 != nil {
		if p1.Val == node.Val {
			if p1.Next == nil {
				p1.Val = 0
				p1 = nil
				continue
			} else {
				p1.Val = p1.Next.Val
				p1.Next = p1.Next.Next
			}
		}

		p1 = p1.Next
	}

	return head
}

func TestDeleteNode(test *testing.T) {
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
	l1 := deleteNode(n7, &ListNode{Val: 6})
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
	l2 := deleteNode(m3, &ListNode{Val: 1})
	for l2 != nil {
		fmt.Println(l2.Val)

		l2 = l2.Next
	}
}

// https://leetcode.cn/problems/shan-chu-lian-biao-de-jie-dian-lcof/

// 删除节点

func deleteNode2(head *ListNode, val int) *ListNode {
	cur := head
	var prev *ListNode
	for cur != nil {
		if cur.Val == val {
			if cur.Next == nil {
				prev.Next = nil
			} else {
				cur.Next = cur.Next.Next
			}
		}

		prev = cur
		cur = cur.Next
	}

	return head
}
