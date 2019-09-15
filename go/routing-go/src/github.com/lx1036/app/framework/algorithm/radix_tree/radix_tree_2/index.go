package radix_tree_2

func longestPrefix(k1 string, k2 string) int  {
	max := len(k1)
	if l := len(k2); l < max {
		max = l
	}

	var i int
	for i = 0; i < max; i++ {
		if k1[i] != k2[i] {
			break
		}
	}

	return i
}


/*func (tree *Tree) Min() (string, interface{}, bool)  {

}

func (tree *Tree) Max() (string, interface{}, bool)  {

}*/



func New() *Tree  {
	return NewFromMap(nil)
}

func NewFromMap(dictionary map[string]interface{}) *Tree  {
	tree := &Tree{root: &node{}}
	for key, value := range dictionary {
		tree.Insert(key, value)
	}

	return tree
}

func recursiveWalk(node *node, fn WalkFn) bool  {
	if node.leaf != nil && fn(node.leaf.key, node.leaf.value) {
		return true
	}

	for _, edge := range node.edges {
		if recursiveWalk(edge.node, fn) {
			return true
		}
	}

	return false
}
