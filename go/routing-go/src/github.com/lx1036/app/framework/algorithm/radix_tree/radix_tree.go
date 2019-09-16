package radix_tree

type Node struct {
	children [16]*Node
	data []byte
}

type KVStore interface {
	Insert(k, b []byte) Trie
	Search(b []byte) []byte
}

func search(node *Node, data []byte) []byte  {
	if node == nil {
		return nil
	} else if len(data) == 0 {
		return node.data
	} else {
		return search(node.children[data[0]], data[1:])
	}
}

func (node *Node) Search(data []byte) []byte {
	return search(node, data)
}

func (node *Node)copy() *Node {
	out := Node{
		children: node.children,
		data:     make([]byte, len(node.data)),
	}
	copy(out.data, node.data)
	return &out
}

func insert(node *Node, key []byte, value []byte) *Node  {
	if node == nil {
		out := Node{}
		return insert(&out, key, value)
	} else if len(key) == 0 {
		out := node.copy()
		out.data = value
		return out
	} else {
		out := node.copy()
		out.children[key[0]] = insert(out.children[key[0]], key[1:], value)
		return out
	}
}

func (node *Node) Insert(key, value []byte) KVStore   {
	return insert(node, key, value)
}

func New() *Node  {
	return nil
}

