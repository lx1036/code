package dfs_bfs_bs

import (
	"github.com/stretchr/testify/assert"
	"k8s-lx1036/k8s/algo/leetcode/tree/binary_search_tree"
	"math"
	"testing"
)

// 深度：二叉树的深度为根节点到最远叶子节点的最长路径上的节点数

// dfs 深度优先搜索
func dfs_104(root *binary_search_tree.Node) int {
	if root == nil {
		return 0
	}

	L := dfs_104(root.Left)  // 左子树深度
	R := dfs_104(root.Right) // 右子树深度
	return int(math.Max(float64(L), float64(R))) + 1
}

// bfs 广度优先搜索
func bfs_104(root *binary_search_tree.Node) int {
	if root == nil {
		return 0
	}

	result := 0
	var queue []*binary_search_tree.Node
	queue = append(queue, root)
	for len(queue) > 0 {
		l := len(queue)
		for l > 0 {
			node := queue[0]
			queue = queue[1:]
			if node.Left != nil {
				queue = append(queue, node.Left)
			}
			if node.Right != nil {
				queue = append(queue, node.Right)
			}

			l--
		}

		result++
	}

	return result
}

func TestDFS_104(test *testing.T) {
	root := binary_search_tree.NewBinarySearchTree()
	assert.Equal(test, 4, bfs_104(root))
}

func TestBFS_104(test *testing.T) {
	root := binary_search_tree.NewBinarySearchTree()
	assert.Equal(test, 4, bfs_104(root))
}
