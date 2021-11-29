package bplustree

type BTree struct {
	root   *branchNode
	first  *leafNode
	leaf   int
	branch int
	height int
}

// 初始化一个 degree=2 的 b+tree
func New() *BTree {
	leaf := newLeafNode(nil)
	branch := newBranchNode(nil, leaf)
	leaf.p = branch

	return &BTree{
		root:   branch,
		first:  leaf,
		leaf:   1,
		branch: 1,
		height: 2,
	}
}

// Insert inserts a (key, value) into the B+ tree
func (tree *BTree) Insert(key int, value string) {
	_, oldIndex, leaf := search(tree.root, key)
	branch := leaf.parent()

	mid, split := leaf.insert(key, value)
	if !split {
		return
	}

	var midNode node
	midNode = leaf
	branch.kcs[oldIndex].child = leaf.next
	leaf.next.setParent(branch)
	interior, interiorP := branch, branch.parent()

	for {
		var oldIndex int
		var newNode *branchNode

		isRoot := interiorP == nil
		if !isRoot {
			oldIndex, _ = interiorP.find(key)
		}

		mid, newNode, split = interior.insert(mid, midNode)
		if !split {
			return
		}

		if !isRoot {
			interiorP.kcs[oldIndex].child = newNode
			newNode.setParent(interiorP)

			midNode = interior
		} else {
			tree.root = newBranchNode(nil, newNode)
			newNode.setParent(tree.root)

			tree.root.insert(mid, interior)
			return
		}

		interior, interiorP = interiorP, interiorP.parent()
	}
}

// Search searches the key in B+ tree
// If the key exists, it returns the value of key and true
// If the key does not exist, it returns an empty string and false
func (tree *BTree) Search(key int) (string, bool) {
	kv, _, _ := search(tree.root, key)
	if kv == nil {
		return "", false
	}
	return kv.value, true
}

func search(n node, key int) (*kv, int, *leafNode) {
	curr := n
	oldIndex := -1

	for {
		switch t := curr.(type) {
		case *leafNode:
			i, ok := t.find(key)
			if !ok {
				return nil, oldIndex, t
			}
			return &t.kvs[i], oldIndex, t
		case *branchNode:
			i, _ := t.find(key)
			curr = t.kcs[i].child
			oldIndex = i
		default:
			panic("")
		}
	}
}
