package linked_list

// INFO: 单链表删除第 n 节点

type ListNode019 struct {
	Val  int
	Next *ListNode019
}

func getListLen(head *ListNode019) (l int) {
	for ; head != nil; head = head.Next {
		l++
	}
	return
}

func removeNthFromEnd(head *ListNode019, n int) *ListNode019 {
	l := getListLen(head)
	prev := *head
	for i := 1; i <= l-n-1; i++ {
		prev = *prev.Next
	}
	cur := prev.Next
	prev.Next = cur.Next

	return head
}

func reverseList019(head *ListNode019) *ListNode019 {
	var prev *ListNode019
	cur := head
	for cur != nil {
		next := cur.Next
		cur.Next = prev
		prev = cur
		cur = next
	}

	return prev
}
