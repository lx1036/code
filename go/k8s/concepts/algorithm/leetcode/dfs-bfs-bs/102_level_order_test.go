package dfs_bfs_bs

import (
	"fmt"
	"k8s-lx1036/k8s/concepts/algorithm/leetcode/tree/binary_search_tree"
	"testing"
)

// https://leetcode-cn.com/problems/binary-tree-level-order-traversal/solution/bfs-de-shi-yong-chang-jing-zong-jie-ceng-xu-bian-l/
func levelOrder(root *binary_search_tree.Node) [][]int {
	if root == nil {
		return nil
	}

	var result [][]int
	var queue []*binary_search_tree.Node
	queue = append(queue, root)
	for len(queue) != 0 {
		var tmp []int
		l := len(queue)
		for i := 0; i < l; i++ {
			node := queue[0]
			queue = queue[1:]

			tmp = append(tmp, node.Value)

			if node.Left != nil {
				queue = append(queue, node.Left)
			}
			if node.Right != nil {
				queue = append(queue, node.Right)
			}
		}

		result = append(result, tmp)
	}

	return result
}

func TestLevelorder(test *testing.T) {
	root := binary_search_tree.NewBinarySearchTree()
	fmt.Println(levelOrder(root))
}
