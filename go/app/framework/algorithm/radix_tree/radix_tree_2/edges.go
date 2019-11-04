package radix_tree_2

import "sort"

type edges []edge

type edge struct {
	label byte
	node *node
}

func (edges edges) Len() int {
	return len(edges)
}

func (edges edges) Less(i, j int) bool {
	return edges[i].label < edges[j].label
}

func (edges edges) Swap(i, j int) {
	edges[i], edges[j] = edges[j], edges[i]
}

func (edges edges)Sort()  {
	sort.Sort(edges)
}
