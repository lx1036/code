

# heap/priority queue
wiki: https://leetcode-cn.com/tag/heap/，
也可以参考这篇文章：**[Kubernetes 源码学习之延时队列](https://mp.weixin.qq.com/s/aZC2MXQFuUu00TgmOozwKQ)**
或者golang 代码：/usr/local/go/src/container/heap/heap.go

A heap is a tree with the property that each node is the minimum-valued node in its subtree.
The minimum element in the tree is the root, at index 0

堆（Heap）是一个可以被看成近似完全二叉树的数组。树上的每一个结点对应数组的一个元素。除了最底层外，该树是完全充满的，而且是从左到右填充。
堆是一种经过排序的完全二叉树，其中任一非终端节点的数据值均不大于（或不小于）其左孩子和右孩子节点的值。 
堆包括最大堆和最小堆：最大堆的每一个节点（除了根结点）的值不大于其父节点；最小堆的每一个节点（除了根结点）的值不小于其父节点。
堆结构的一个常见应用是建立优先队列（Priority Queue）。

> 最小堆对应的完全二叉树中所有结点的值均不大于其左右子结点的值，且一个结点与其兄弟之间没有必然的联系

完全二叉树：一棵深度为k的有n个结点的二叉树，对树中的结点按从上至下、从左到右的顺序进行编号，如果编号为i（1≤i≤n）的结点与满二叉树中编号为i的结点在二叉树中的位置相同，则这棵二叉树称为完全二叉树。
比如：数组 [1,2,3,4,5,6,7,8,9,10]，组成的完全二叉树是：
                1
            2       3
          4   5   6   7
        8  9 10
> 位置i的节点的父节点位置一定是(i-1)/2，两个子节点一定是2i+1和2i+2

heap最核心的两个方法是上浮和下沉：i节点比父节点大，需要交换这两个节点，交换后可能比新的父节点大，递归持续交换，称为上浮；i节点比父节点小，不断下沉。
