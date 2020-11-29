package binary_search_tree

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFind(test *testing.T) {
	root := NewBinarySearchTree()
	node := root.find(9)
	assert.Equal(test, node.Value, 9)

	node = root.find(2)
	assert.Nil(test, node)
}

func TestInsert(test *testing.T) {
	root := NewBinarySearchTree()
	err := root.insert(13)
	assert.Nil(test, err)

	node := root.find(13)
	assert.Equal(test, node.Value, 13)
}

func TestOrder(test *testing.T) {
	root := NewBinarySearchTree()
	root.Inorder()
	fmt.Println()
	root.Preorder()
	fmt.Println()
	root.Postorder()
}

func TestMinMax(test *testing.T) {
	root := NewBinarySearchTree()
	min := root.Min()
	assert.Equal(test, min.Value, 4)

	max := root.Max()
	assert.Equal(test, max.Value, 20)
}
