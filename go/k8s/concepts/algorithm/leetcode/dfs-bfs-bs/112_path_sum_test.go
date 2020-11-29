package dfs_bfs_bs

import (
	"github.com/stretchr/testify/assert"
	"k8s-lx1036/k8s/concepts/algorithm/leetcode/tree/binary_search_tree"
	"testing"
)

// https://leetcode-cn.com/problems/path-sum/

func hasPathSum(root *binary_search_tree.Node, sum int) bool {
	return dfs_112(root, sum)
}

// 时间复杂度O(n)，空间复杂度O(logN)
func dfs_112(root *binary_search_tree.Node, sum int) bool {
	if root == nil {
		return false
	}

	if root.Left == nil && root.Right == nil {
		return root.Value == sum
	}

	return dfs_112(root.Left, sum-root.Value) || dfs_112(root.Right, sum-root.Value)
}

// 还没完成？？
func bfs(root *binary_search_tree.Node, sum int) bool {
	if root == nil {
		return 0 == sum
	}

	var queue []*binary_search_tree.Node
	queue = append(queue, root)
	var tmp []int
	for len(queue) != 0 {
		l := len(queue)
		for i := 0; i < l; i++ {
			node := queue[0]
			queue = queue[1:]

			if node.Left != nil {
				queue = append(queue, node.Left)
				tmp = append(tmp, node.Value)
			}
			if node.Right != nil {
				queue = append(queue, node.Right)
			}
		}

		//tmp = append(tmp, tmp1)
	}

	return false
}

func TestPathSum(test *testing.T) {
	root := binary_search_tree.NewBinarySearchTree()
	result := hasPathSum(root, 27)
	assert.Equal(test, true, result)
}
