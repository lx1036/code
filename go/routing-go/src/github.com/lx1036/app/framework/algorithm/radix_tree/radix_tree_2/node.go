package radix_tree_2

import "sort"

type node struct {
	leaf *leafNode
	prefix string
	edges edges
}

type leafNode struct {
	key string
	value interface{}
}

func (node *node) isLeaf() bool  {
	return node.leaf != nil
}

func (node *node) getEdge(label byte) *node  {
	number := len(node.edges)
	idx := sort.Search(number, func(i int) bool {
		return node.edges[i].label >= label
	})

	if idx < number && node.edges[idx].label == label {
		return node.edges[idx].node
	}

	return nil
}

func (node *node) addEdge(edge edge)  {
	node.edges = append(node.edges, edge)
	node.edges.Sort()
}

func (node *node) updateEdge(label byte, newNode *node)  {
	number := len(node.edges)
	idx := sort.Search(number, func(i int) bool {
		return node.edges[i].label >= label
	})

	if idx < number && node.edges[idx].label == label {
		node.edges[idx].node = newNode
		return
	}

	panic("missing edge")
}
