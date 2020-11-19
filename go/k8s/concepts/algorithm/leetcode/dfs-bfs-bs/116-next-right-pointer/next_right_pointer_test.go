package _16_next_right_pointer

import (
	"fmt"
	"testing"
)

// https://leetcode-cn.com/problems/populating-next-right-pointers-in-each-node/solution/

type Node struct {
	Val   int
	Left  *Node
	Right *Node
	Next  *Node
}

func connect(root *Node) *Node {
	return bfs(root)
}

func bfs(root *Node) *Node {
	if root == nil {
		return root
	}
	
	var queue []*Node
	queue = append(queue, root)
	for len(queue) != 0 {
		l := len(queue)
		for i := 0; i < l; i++ {
			node := queue[0]
			queue = queue[1:]
			if i < l-1 {
				node.Next = queue[0]
			}
			
			if node.Left != nil {
				queue = append(queue, node.Left)
			}
			if node.Right != nil {
				queue = append(queue, node.Right)
			}
		}
	}
	
	return root
}


func TestConnect(test *testing.T) {

}

func TestName(test *testing.T) {
	type Person struct {
		Name string
	}
	
	p := &Person{Name: "test"}
	q := p
	q.Name = "test2"
	fmt.Println(p.Name) // test2
	
	var queue []*Person
	queue = append(queue, p)
	node := queue[0]
	queue = queue[1:]
	fmt.Println(node.Name, queue)
}
