package binary_search_tree

import "fmt"

// Binary Search Tree

type Node struct {
	Value int
	Left *Node
	Right *Node
}

// 深度优先搜索 Depth-First-Search DFS
// https://leetcode-cn.com/problems/diameter-of-binary-tree/solution/er-cha-shu-de-zhi-jing-by-leetcode-solution/
func (node *Node) Diameter() int {

	return 0
}

func (node *Node) find(value int) *Node {
	current := node
	for current != nil {
		if value > current.Value {
			current = current.Right
		} else if value < current.Value {
			current = current.Left
		} else if value == current.Value {
			return current
		}
	}
	
	return nil
}

// 递归实现插入节点，很巧妙
// 数据大于当前节点值，插入右节点；数据小于当前节点值，插入左节点
func (node *Node) insert(value int) error {
	if value == node.Value {
		return nil
	}
	
	if value > node.Value {
		if node.Right == nil {
			node.Right = &Node{Value: value}
		} else {
			node.Right.insert(value)
		}
	} else if value < node.Value {
		if node.Left == nil {
			node.Left = &Node{Value: value}
		} else {
			node.Left.insert(value)
		}
	}
	
	return nil
}

// 还未完成？？
func (node *Node) delete(value int) error {
	if value == node.Value {
		if node.Left == nil && node.Right == nil {
			
		}
	}
	
	if value > node.Value {
		node.Right.delete(value)
	} else if value < node.Value {
		node.Left.delete(value)
	}
	
	
	return nil
}

// 遍历树：根据特定顺序遍历树的每一个节点
// 中序遍历：左子树 -> 根节点 -> 右子树
// 前序遍历：根节点 -> 左子树 -> 右子树
// 后序遍历：左子树 -> 右子树 -> 根节点

func (node *Node) MiddleOrder()  {
	current := node
	if current != nil {
		current.Left.MiddleOrder()
		fmt.Println(current.Value)
		current.Right.MiddleOrder()
	}
}

func (node *Node) PreOrder()  {
	current := node
	if current != nil {
		fmt.Println(current.Value)
		current.Left.MiddleOrder()
		current.Right.MiddleOrder()
	}
}

func (node *Node) PostOrder()  {
	current := node
	if current != nil {
		current.Left.MiddleOrder()
		current.Right.MiddleOrder()
		fmt.Println(current.Value)
	}
}

func (node *Node) Min() *Node {
	current := node
	if current.Left == nil {
		return current
	}
	
	if current != nil {
		return current.Left.Min()
	}
	
	return nil
}

func (node *Node) Max() *Node{
	current := node
	if current.Left == nil {
		return current
	}
	
	if current != nil {
		return current.Right.Max()
	}
	
	return nil
}


//          10
//        /    \
//       8      15
//      /  \   /  \
//     4    9 11  20
//      \
//       5

func NewBinarySearchTree() *Node {
	/*node41 := &Node{Value: 5}
	node31 := &Node{Value: 4, right: node41}
	node32 := &Node{Value: 9}
	node21 := &Node{Value: 8, left: node31, right: node32}
	node33 := &Node{Value: 11}
	node34 := &Node{Value: 20}
	node22 := &Node{Value: 15, left: node33, right: node34}
	root := &Node{Value: 10, left: node21, right: node22}*/

	root := &Node{Value: 10}
	root.insert(8)
	root.insert(15)
	root.insert(4)
	root.insert(9)
	root.insert(11)
	root.insert(20)
	root.insert(5)

	return root
}
