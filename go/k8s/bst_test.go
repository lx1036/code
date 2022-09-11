package k8s

import "fmt"

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// 翻转二叉树
// https://leetcode.cn/problems/invert-binary-tree/solution/dong-hua-yan-shi-liang-chong-shi-xian-226-fan-zhua/
// 递归：交换左右节点，再递归交换左右子节点
func invertTree(root *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}
	left := invertTree(root.Left)
	right := invertTree(root.Right)
	root.Left = right
	root.Right = left
	return root
}

// 遍历树：根据特定顺序遍历树的每一个节点
// 中序遍历：左子树 -> 根节点 -> 右子树
// 前序遍历：根节点 -> 左子树 -> 右子树
// 后序遍历：左子树 -> 右子树 -> 根节点

// 前序遍历：根节点 -> 左子树 -> 右子树
func (node *TreeNode) Preorder() {
	current := node
	if current != nil {
		fmt.Println(current.Val)
		current.Left.Preorder()
		current.Right.Preorder()
	}
}

// 后序遍历：左子树 -> 右子树 -> 根节点
func (node *TreeNode) Postorder() {
	current := node
	if current != nil {
		current.Left.Postorder()
		current.Right.Postorder()
		fmt.Println(current.Val)
	}
}

// 中序遍历：左子树 -> 根节点 -> 右子树
func (node *TreeNode) Inorder() {
	current := node
	if current != nil {
		current.Left.Inorder()
		fmt.Println(current.Val)
		current.Right.Inorder()
	}
}
