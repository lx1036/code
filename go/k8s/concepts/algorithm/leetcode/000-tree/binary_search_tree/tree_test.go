package binary_search_tree

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)


//          10
//        /    \
//       8      15
//      /  \   /  \
//     4    9 11  20
//      \
//       5
func NewBinarySearchTree() *Node {
	/*node41 := &Node{value: 5}
	node31 := &Node{value: 4, right: node41}
	node32 := &Node{value: 9}
	node21 := &Node{value: 8, left: node31, right: node32}
	node33 := &Node{value: 11}
	node34 := &Node{value: 20}
	node22 := &Node{value: 15, left: node33, right: node34}
	root := &Node{value: 10, left: node21, right: node22}*/
	
	root := &Node{value: 10}
	root.insert(8)
	root.insert(15)
	root.insert(4)
	root.insert(9)
	root.insert(11)
	root.insert(20)
	root.insert(5)
	
	return root
}

func TestFind(test *testing.T) {
	root := NewBinarySearchTree()
	node := root.find(9)
	assert.Equal(test, node.value, 9)
	
	node = root.find(2)
	assert.Nil(test, node)
}

func TestInsert(test *testing.T) {
	root := NewBinarySearchTree()
	err := root.insert(13)
	assert.Nil(test, err)
	
	node := root.find(13)
	assert.Equal(test, node.value, 13)
}

func TestOrder(test *testing.T) {
	root := NewBinarySearchTree()
	root.MiddleOrder()
	fmt.Println()
	root.PreOrder()
	fmt.Println()
	root.PostOrder()
}

func TestMinMax(test *testing.T) {
	root := NewBinarySearchTree()
	min := root.Min()
	assert.Equal(test, min.value, 4)
	
	max := root.Max()
	assert.Equal(test, max.value, 20)
}
