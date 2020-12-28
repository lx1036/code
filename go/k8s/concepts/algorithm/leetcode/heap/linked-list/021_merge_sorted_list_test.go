package linked_list

import (
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/merge-two-sorted-lists/

// 合并两个升序链表

// 迭代，时间复杂度O(n+m)
func mergeTwoLists(l1 *ListNode, l2 *ListNode) *ListNode {
	root1 := l1
	root2 := l2
	root := &ListNode{}
	prev := root
	// 只是迭代两个链表的min(root1, root2)长度
	for root1 != nil && root2 != nil {
		if root1.Val <= root2.Val {
			prev.Next = root1
			root1 = root1.Next
		} else {
			prev.Next = root2
			root2 = root2.Next
		}

		prev = prev.Next
	}

	// 谁先迭代完成，后面节点直接合并就行，不用迭代
	if root1 == nil {
		prev.Next = root2
	} else {
		prev.Next = root1
	}

	return root.Next
}

func TestMergeTwoLists(test *testing.T) {
	a1 := &ListNode{Val: 5, Next: nil}
	a2 := &ListNode{Val: 4, Next: a1}
	a3 := &ListNode{Val: 1, Next: a2}

	b0 := &ListNode{Val: 6, Next: nil}
	b1 := &ListNode{Val: 3, Next: b0}
	b2 := &ListNode{Val: 2, Next: b1}
	b3 := &ListNode{Val: 1, Next: b2}

	c := mergeTwoLists(a3, b3)
	tmp := c
	for tmp != nil {
		fmt.Println(tmp.Val)
		tmp = tmp.Next
	}
}
