package _43_diameter_binary_tree

import (
	"github.com/stretchr/testify/assert"
	"k8s-lx1036/k8s/concepts/algorithm/leetcode/tree/binary_search_tree"
	"math"
	"testing"
)

// https://leetcode-cn.com/problems/diameter-of-binary-tree/

var result = float64(1)

// 二叉树直径
func diameterOfBinaryTree(root *binary_search_tree.Node) int {
	depth(root)
	return int(result) - 1
}

func depth(node *binary_search_tree.Node) int {
	if node == nil {
		return 0
	}

	L := depth(node.Left)  // 左子树深度
	R := depth(node.Right) // 右子树深度

	result = math.Max(result, float64(L+R+1))
	return int(math.Max(float64(L), float64(R))) + 1
}

func TestDiameter(test *testing.T) {
	root := binary_search_tree.NewBinarySearchTree()
	assert.Equal(test, 5, diameterOfBinaryTree(root))
}
