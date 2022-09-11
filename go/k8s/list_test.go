package k8s

// (1) 反转单链表
type List2 struct {
	value int
	next  *List2
}

func (l *List2) reverse() *List2 {
	cur := l
	var prev *List2
	for cur != nil {
		next := cur.next
		cur.next = prev
		prev = cur
		cur = next
	}
	return prev
}
