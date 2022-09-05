package k8s

/*

# btree
btree 优点：树的高度很低，适合磁盘存储时IO很少，存储数据量却很大。且支持范围查询，有序性查询。

对于一个阶数 degree 为 m 的 B-Tree，定义如下：
(1)每个结点最多有 m 个子结点；
(2)每个非叶子结点（根结点除外）至少含有 m/2 个子结点；
(3)如果根结点不是叶子结点，那么根结点至少有两个子结点；
(4)对于一个非叶子结点而言，它最多能存储 m-1 个数据；(那非叶子节点，子节点有 m/2 < k < m-1)
(5)每个节点上，所有的数据都是有序的，从左至右，依次从小到大排列；
(6)每个节点的左子树的值均小于当前节点值，右子树的值均大于当前节点值；
(7)每个节点都存有索引和数据(这里索引和数据都存储在节点中)


插入和查找，可以看动画过程，非常形象：https://www.cs.usfca.edu/~galles/visualization/BTree.html
依次插入：50、30、40、25、15、10、13、18、60、55、45、26、17、8、3、5

*/
