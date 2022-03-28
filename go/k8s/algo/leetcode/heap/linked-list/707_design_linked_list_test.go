package linked_list

import (
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/design-linked-list/solution/she-ji-lian-biao-by-leetcode/
// 链表所有节点都是0 - index，index起点0开始

type MyLinkedList struct {
	Value  int
	Next   *MyLinkedList
	Length int
}

func Constructor() MyLinkedList {
	return MyLinkedList{}
}

func (this *MyLinkedList) Get(index int) int {
	//if index < 0 {
	if index >= this.Length || index < 0 {
		return -1
	}

	current := this
	for i := 0; i < index; i++ {
		current = current.Next
	}

	return current.Value
}

func (this *MyLinkedList) AddAtHead(val int) {

	current := &MyLinkedList{Value: val, Length: this.Length + 1}
	current.Next = this
	this = current
}

func (this *MyLinkedList) AddAtTail(val int) {

}

func (this *MyLinkedList) AddAtIndex(index int, val int) {

	if index > this.Length {
		return
	}
	if index < 0 {
		index = 0
	}

	node := &MyLinkedList{Value: val, Length: this.Length + 1}

	current := this
	for i := 0; i < index-1; i++ { // 获取当前位置的上一个节点
		current = current.Next
	}

	next := current.Next
	current.Next = node
	node.Next = next
}

func (this *MyLinkedList) DeleteAtIndex(index int) {

}

func TestLinkedList(test *testing.T) {
	//obj := Constructor()
	//param_1 := obj.Get(index)
	//obj.AddAtHead(val)
	//obj.AddAtTail(val)
	//obj.AddAtIndex(index,val)
	//obj.DeleteAtIndex(index)

	list1 := &MyLinkedList{Value: 1, Length: 0}
	head := &MyLinkedList{Value: 2, Length: 1, Next: list1}
	current := head
	for current != nil {
		fmt.Println(current.Value)
		current = current.Next
	}

	fmt.Println(head.Value)
}
