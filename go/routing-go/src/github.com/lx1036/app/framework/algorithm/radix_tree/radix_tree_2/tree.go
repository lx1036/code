package radix_tree_2

import "strings"

type Tree struct {
	root *node
	size int
}

type WalkFn func(key string, value interface{}) bool

func (tree *Tree) Insert(key string, value interface{}) (interface{}, bool)  {
	var parent *node
	search := key
	rootNode := tree.root

	for  {
		if len(search) == 0 {
			if rootNode.isLeaf() {
				old := rootNode.leaf.value
				rootNode.leaf.value = value
				return old, true
			}

			rootNode.leaf = &leafNode{
				key:search,
				value:value,
			}
			tree.size++
			return nil, false
		}

		parent = rootNode
		rootNode = rootNode.getEdge(search[0])
		if rootNode == nil { // No edge, create one
			edge := edge{
				label:search[0],
				node:&node{
					leaf:&leafNode{
						key:search,
						value:value,
					},
					prefix: search,
				},
			}
			parent.addEdge(edge)
			tree.size++
			return nil, false
		}

		commonPrefix := longestPrefix(search, rootNode.prefix)
		if commonPrefix == len(rootNode.prefix) {
			search = search[commonPrefix:]
			continue
		}

		// Split the node
		tree.size++
		child := &node{
			prefix:search[:commonPrefix],
		}
		parent.updateEdge(search[0], child)

		// Restore the existing node
		child.addEdge(edge{
			label:rootNode.prefix[commonPrefix],
			node:rootNode,
		})
		rootNode.prefix = rootNode.prefix[commonPrefix:]

		// if the new key is a subnet, add it into this node
		leaf := &leafNode{
			key: search,
			value:value,
		}
		search = search[commonPrefix:]
		if len(search) == 0 {
			child.leaf = leaf
			return nil, false
		}
		child.addEdge(edge{
			label:search[0],
			node:&node{
				leaf:leaf,
				prefix:search,
			},
		})
		return nil, false
	}
}

func (tree *Tree) Get(key string) (interface{}, bool)  {
	node := tree.root
	search := key
	for  {
		if len(search) == 0 {
			if node.isLeaf() {
				return node.leaf.value, true
			}
		}

		node = node.getEdge(search[0])
		if node == nil {
			break
		}
		if strings.HasPrefix(search, node.prefix) {
			search = search[len(node.prefix):]
		} else {
			break
		}
	}

	return nil, false
}

func (tree *Tree) Delete(key string) (interface{}, bool) {
	search := key
	rootNode := tree.root
	var parent *node
	for {
		if len(search) == 0 {
			if !rootNode.isLeaf() {
				break
			}
			goto DELETE
		}

		parent = rootNode
		rootNode = rootNode.getEdge(search[0])
		if rootNode == nil {
			break
		}
		if strings.HasPrefix(search, rootNode.prefix) {
			search = search[len(rootNode.prefix):]
		} else {
			break
		}
	}

	return nil, false

	DELETE:
		leaf := rootNode.leaf
		rootNode.leaf = nil
		tree.size--
		if parent != nil && len(rootNode.edges) == 0 {

		}

		// Check if delete this node from the parent

		// Check if merge this node

		// Check if merge the parent's other node
		if parent != nil {

		}

		return leaf.value, true
}

func (tree *Tree) Len() int  {
	return tree.size
}

func (tree *Tree) Walk(fn WalkFn)  {
	recursiveWalk(tree.root, fn)
}
