package leetcode

// https://leetcode.cn/problems/kth-node-from-end-of-list-lcci/

type ListNode struct {
	Val  int
	Next *ListNode
}

func kthToLast(head *ListNode, k int) int {
	l := reverseList(head)
	var next *ListNode
	for i := 0; i < k; i++ {
		next = l
		l = l.Next
	}

	return next.Val
}

func reverseList(l *ListNode) *ListNode {
	cur := l
	var prev *ListNode
	for cur != nil {
		next := cur.Next
		cur.Next = prev
		prev = cur
		cur = next
	}

	return prev
}
