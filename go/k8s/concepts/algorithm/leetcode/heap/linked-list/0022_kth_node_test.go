package linked_list

import (
	"fmt"
	"testing"
)

// 剑指 Offer 22. 链表中倒数第k个节点
// https://leetcode-cn.com/problems/lian-biao-zhong-dao-shu-di-kge-jie-dian-lcof/

func reverse(head *ListNode) *ListNode {
	if head.Next == nil {
		return head
	}

	prev := head
	current := head.Next
	for current != nil {
		next := current.Next
		current.Next = prev
		prev = current
		current = next
	}

	return prev
}

func getKthFromEnd(head *ListNode, k int) *ListNode {
	root := head
	r := reverse(root)

	tmp := r

	var knode *ListNode
	for i := 0; i < k; i++ {
		knode = tmp
		tmp = tmp.Next
	}

	if knode == nil {
		return nil
	}

	knode.Next = nil

	return reverse(r)

	//return reverse(r, knode)
}

func TestGetKthFromEnd(test *testing.T) {
	n1 := &ListNode{
		Val: 5,
	}
	n2 := &ListNode{
		Val:  4,
		Next: n1,
	}
	n3 := &ListNode{
		Val:  3,
		Next: n2,
	}
	n4 := &ListNode{
		Val:  2,
		Next: n3,
	}
	n5 := &ListNode{
		Val:  1,
		Next: n4,
	}

	list := getKthFromEnd(n5, 2)
	current := list
	for current != nil {
		fmt.Println(current.Val)
		current = current.Next
	}

}
