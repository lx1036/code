package tree

import (
    "github.com/sirupsen/logrus"
    "testing"
)

type Node struct {
    value int
    left  *Node
    right *Node
}

// https://juejin.cn/post/6844903503807119374

// PreOrder 先序：根左右
func traverseTreeStackPreOrder(root *Node) []int {
    var results []int
    stack := make([]*Node, 0)
    for root != nil || len(stack) != 0 {
        for root != nil {
            results = append(results, root.value)
            stack = append(stack, root)
            root = root.left
        }

        root = stack[len(stack)-1]
        stack = stack[:len(stack)-1]
        root = root.right
    }

    return results
}

// InOrder 中序：左根右
func traverseTreeStackInOrder(root *Node) []int {
    var results []int
    stack := make([]*Node, 0)
    for root != nil || len(stack) != 0 {
        for root != nil {
            stack = append(stack, root)
            root = root.left
        }

        root = stack[len(stack)-1]
        stack = stack[:len(stack)-1]
        results = append(results, root.value)
        root = root.right
    }

    return results
}

// PostOrder 后序：左右根
func traverseTreeStackPostOrder(root *Node) []int {
    var results []int
    stack := make([]*Node, 0)
    var last *Node
    for root != nil || len(stack) != 0 {
        for root != nil {
            stack = append(stack, root)
            root = root.left
        }

        root = stack[0]
        if root.right == nil || root.right == last {
            results = append(results, root.value)
            // pop
            root = stack[len(stack)-1]
            stack = stack[:len(stack)-1]
            // 记录上一个访问的节点，用于判断"访问根节点之前，右子树是否已经访问过"
            last = root
            // 表示不需要转向，继续弹栈
            root = nil
        } else {
            root = root.right
        }
    }

    return results
}

func TestTraverseTree(test *testing.T) {
    //     1
    //    / \
    //   2    5
    //  / \   / \
    // 3   4  6  7
    root := &Node{
        value: 1,
        left: &Node{
            value: 2,
            left: &Node{
                value: 3,
            },
            right: &Node{
                value: 4,
            },
        },
        right: &Node{
            value: 5,
            left: &Node{
                value: 6,
            },
            right: &Node{
                value: 7,
            },
        },
    }

    results := traverseTreeStackPreOrder(root)
    logrus.Infof("PreOrder: %+v", results) // [1 2 3 4 5 6 7]

    results = traverseTreeStackInOrder(root)
    logrus.Infof("InOrder: %+v", results) // [3 2 4 1 6 5 7]

    results = traverseTreeStackPostOrder(root)
    logrus.Infof("PostOrder: %+v", results) // ??? 没有成功
}
