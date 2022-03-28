package linked_list

import (
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/odd-even-linked-list/

//字节面试题
//[1]链表，奇数位置按序增长，偶数位置按序递减，如何能实现链表从小到大？（2020.10 字节跳动-后端）
//[2]奇偶生序倒序链表的重新排序组合，例如：18365472（2020.08 字节跳动-后端）
//[3]1->4->3->2->5 给定一个链表奇数部分递增，偶数部分递减，要求在O(n)时间复杂度内将链表变成递增，5分钟左右（2020.07 字节跳动-测试开发）
//[4]奇数位升序偶数位降序的链表要求时间O(n)空间O(1)的排序？(2020.07 字节跳动-后端)

// 字节跳动高频题之排序奇升偶降链表: https://zhuanlan.zhihu.com/p/311113031

// 迭代
func oddEvenList(head *ListNode) *ListNode {

	root := &ListNode{}
	prev := root
	current := head
	for current != nil && current.Next != nil {
		next := current.Next.Next

		prev.Next = current
		//prev.Next = next

		current = next
		prev = prev.Next
	}

	return root.Next
}

func TestOddEvenLinkedList(test *testing.T) {
	a1 := &ListNode{Val: 7, Next: nil}
	a2 := &ListNode{Val: 4, Next: a1}
	a3 := &ListNode{Val: 6, Next: a2}
	a4 := &ListNode{Val: 5, Next: a3}
	a5 := &ListNode{Val: 3, Next: a4}
	a6 := &ListNode{Val: 1, Next: a5}
	a7 := &ListNode{Val: 2, Next: a6}

	c := oddEvenList(a7)
	tmp := c
	for tmp != nil {
		fmt.Println(tmp.Val)
		tmp = tmp.Next
	}
}
