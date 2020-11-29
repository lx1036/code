package dfs_bfs_bs

import "k8s-lx1036/k8s/concepts/algorithm/leetcode/tree/binary_search_tree"

// https://leetcode-cn.com/problems/serialize-and-deserialize-binary-tree/

type Codec struct {
}

func Constructor() Codec {

	return Codec{}
}

// Serializes a tree to a single string.
func (this *Codec) serialize(root *binary_search_tree.Node) string {

	return ""
}

// Deserializes your encoded data to tree.
func (this *Codec) deserialize(data string) *binary_search_tree.Node {

	return nil
}

/**
 * Your Codec object will be instantiated and called as such:
 * ser := Constructor();
 * deser := Constructor();
 * data := ser.serialize(root);
 * ans := deser.deserialize(data);
 */
