package _04_max_depth_bst

import "math"

type TreeNode struct {
	value int
	left *TreeNode
	right *TreeNode
}


// dfs
func maxDepth(root *TreeNode) int {
	if root == nil {
		return 0
	}

	return int(math.Max(float64(maxDepth(root.left)), float64(maxDepth(root.right)))) + 1
}

