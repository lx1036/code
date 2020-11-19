package _01_symmetic_tree

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s-lx1036/k8s/concepts/algorithm/leetcode/tree/binary_search_tree"
	"testing"
)

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func isSymmetric(root *binary_search_tree.Node) bool {
	if root == nil {
		return true
	}

	var queue []*binary_search_tree.Node
	queue = append(queue, root)
	result := true
	for len(queue) != 0 {
		var tmp []int
		l := len(queue)
		for i := 0; i < l; i++ {
			node := queue[0]
			queue = queue[1:]
			if node.Left != nil {
				queue = append(queue, node.Left)
				tmp = append(tmp, node.Value)
			} else {
				tmp = append(tmp, 0)
			}
			if node.Right != nil {
				queue = append(queue, node.Right)
				tmp = append(tmp, node.Value)
			} else {
				tmp = append(tmp, 0)
			}
		}

		if !valid(tmp) {
			result = false
			break
		}
	}

	return result
}

func valid(queue []int) bool {
	if len(queue)%2 != 0 {
		return false
	}

	l := len(queue)
	for i, j := 0, l-1; j >= i; i, j = i+1, j-1 {
		if queue[i] != queue[j] {
			return false
		}
	}
	return true
}

func isSymmetricRecursive(root *binary_search_tree.Node) bool {
	return dfs(root, root)
}

func dfs(node1, node2 *binary_search_tree.Node) bool {
	if node1 == nil && node2 == nil {
		return true
	}
	if node1 == nil || node2 == nil {
		return false
	}

	return node1.Value == node2.Value && dfs(node1.Left, node2.Right) && dfs(node1.Right, node2.Left)
}

func TestSymmetric(test *testing.T) {
	values := []int{1, 2}
	root := binary_search_tree.NewBinarySearchTreeByValues(values)
	result := isSymmetric(root)
	fmt.Println(result)
	assert.Equal(test, false, result)

	result = isSymmetricRecursive(root)
	assert.Equal(test, false, result)

	values = []int{1, 2, 2, 3, 4, 4, 3}
	root = binary_search_tree.NewBinarySearchTreeByValues(values)
	result = isSymmetric(root)
	assert.Equal(test, false, result)

	result = isSymmetricRecursive(root)
	assert.Equal(test, false, result)
}
