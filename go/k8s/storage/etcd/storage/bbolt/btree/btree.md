

# 平衡树b-tree(balance tree)
文档：
**[拜托，别再问我什么是 B+ 树了](https://leetcode-cn.com/circle/article/M2rEuR/)**
**[B树自在人心，看不懂，我当场把这个树吃掉！](https://www.bilibili.com/video/BV1Aa4y1j7a4/)**
**[漫画：什么是B+树？](https://zhuanlan.zhihu.com/p/54102723)**

https://github.com/google/btree

btree 的使用: https://zhengyinyong.com/post/btree-usage/


# btree
动态 btree 过程：https://www.cs.usfca.edu/~galles/visualization/BTree.html
索引数据结构之B-Tree: https://juejin.cn/post/6844904120051056647 , 好文章！！！

对于一个阶数 degree 为 m 的 B-Tree，定义如下：
(1)每个结点最多有 m 个子结点；
(2)每个非叶子结点（根结点除外）至少含有 m/2 个子结点；
(3)如果根结点不是叶子结点，那么根结点至少有两个子结点；
(4)对于一个非叶子结点而言，它最多能存储 m-1 个数据；(那非叶子节点，子节点有 m/2 < k < m-1)
(5)每个节点上，所有的数据都是有序的，从左至右，依次从小到大排列；
(6)每个节点的左子树的值均小于当前节点值，右子树的值均大于当前节点值；
(7)每个节点都存有索引和数据

## 插入insert流程
insert: 向当前结点中插入 Item 后，判断当前结点的 Item 数量是否小于等于 m-1，如果小于，则插入结束；否则需要将当前结点进行分裂，如何分裂呢？
在 m/2 处拆分，形成左右两部分，即两个新的子结点，然后将 m/2 处的 Item 移到父节点当中（从最中间分裂）。




