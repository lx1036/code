

# heap/priority queue
wiki: https://leetcode-cn.com/tag/heap/，
也可以参考这篇文章：https://mp.weixin.qq.com/s/aZC2MXQFuUu00TgmOozwKQ
或者golang 代码：/usr/local/go/src/container/heap/heap.go

A heap is a tree with the property that each node is the minimum-valued node in its subtree.
The minimum element in the tree is the root, at index 0

堆（Heap）是一个可以被看成近似完全二叉树的数组。树上的每一个结点对应数组的一个元素。除了最底层外，该树是完全充满的，而且是从左到右填充。
堆是一种经过排序的完全二叉树，其中任一非终端节点的数据值均不大于（或不小于）其左孩子和右孩子节点的值。 
堆包括最大堆和最小堆：最大堆的每一个节点（除了根结点）的值不大于其父节点；最小堆的每一个节点（除了根结点）的值不小于其父节点。
堆结构的一个常见应用是建立优先队列（Priority Queue）。


