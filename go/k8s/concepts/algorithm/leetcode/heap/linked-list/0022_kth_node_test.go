package linked_list

import (
	"fmt"
	"testing"
)

// 剑指 Offer 22. 链表中倒数第k个节点
// https://leetcode-cn.com/problems/lian-biao-zhong-dao-shu-di-kge-jie-dian-lcof/

// 解法1：或者遍历链表获得长度n，再以n-k+1为头结点

// 解法2：双指针，p2先走k，然后p1、p2一起同速度走，直至p2走到尾结点，p1就是第k个节点
func getKthFromEnd2(head *ListNode, k int) *ListNode {
	p2 := head
	for i := 0; i < k; i++ {
		p2 = p2.Next
	}

	p1 := head
	for p2 != nil {
		p1 = p1.Next
		p2 = p2.Next
	}

	return p1
}

// 解法3：反转链表，p1走k步
func reverse(head *ListNode) *ListNode {
	if head == nil || head.Next == nil {
		return head
	}

	current := head.Next
	prev := head
	prev.Next = nil
	for current != nil {
		next := current.Next
		current.Next = prev
		prev = current
		current = next
	}

	return prev
}

// 两次反转链表
func getKthFromEnd(head *ListNode, k int) *ListNode {
	root := head
	r := reverse(root)

	p1 := r

	// 这里i从1开始计数
	for i := 1; i < k; i++ {
		p1 = p1.Next
	}

	// 这个赋值很重要，设置kth为尾节点
	p1.Next = nil

	return reverse(r)
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

	fmt.Println()

	m1 := &ListNode{
		Val: 5,
	}
	m2 := &ListNode{
		Val:  4,
		Next: m1,
	}
	m3 := &ListNode{
		Val:  3,
		Next: m2,
	}
	m4 := &ListNode{
		Val:  2,
		Next: m3,
	}
	m5 := &ListNode{
		Val:  1,
		Next: m4,
	}
	list2 := getKthFromEnd2(m5, 2)
	current2 := list2
	for current2 != nil {
		fmt.Println(current2.Val)
		current2 = current2.Next
	}
}
