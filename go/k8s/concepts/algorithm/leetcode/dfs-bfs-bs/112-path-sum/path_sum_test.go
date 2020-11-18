package _12_path_sum

import (
	"k8s-lx1036/k8s/concepts/algorithm/leetcode/tree/binary_search_tree"
	"testing"
)

// https://leetcode-cn.com/problems/path-sum/

func hasPathSum(root *binary_search_tree.Node, sum int) bool {

	return true
}

/*func dfs(root *binary_search_tree.Node) int {
	if root == nil {
		return 0
	}

	l := root.Value + dfs(root.Left)
	r := root.Value + dfs(root.Right)

}*/

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

}

func TestPathSum(test *testing.T) {

}


