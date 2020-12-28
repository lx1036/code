package lru

import (
	golist "container/list"
	"fmt"
	"testing"
)

func TestDoublyLinkedList(test *testing.T) {
	// Create a new list and put some numbers in it.
	l := golist.New()
	e4 := l.PushBack(4)
	e1 := l.PushFront(1)
	l.InsertBefore(3, e4)
	l.InsertAfter(2, e1)

	// Iterate through list and print its contents.
	for e := l.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}

	// Output:
	// 1
	// 2
	// 3
	// 4
}
