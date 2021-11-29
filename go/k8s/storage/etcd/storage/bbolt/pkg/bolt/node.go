package bolt

import (
	"bytes"
	"fmt"
	"k8s.io/klog/v2"
	"sort"
	"unsafe"
)

// INFO: 表示一个node里的(key,value)数组元素，同时有pageid
// INFO: 分支节点和叶子节点也复用同一个结构体。B+tree里 branch node 里存的是key，leaf node才会存key-value
type inode struct {
	flags uint32
	pgid  pgid
	key   []byte
	value []byte
}

type inodes []inode

type nodes []*node

func (s nodes) Len() int      { return len(s) }
func (s nodes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s nodes) Less(i, j int) bool {
	return bytes.Compare(s[i].inodes[0].key, s[j].inodes[0].key) == -1 // a < b
}

// INFO: 每个桶对应一棵 B+ 树，命名空间是可以嵌套的，因此 BoltDB 的 Bucket 间也是允许嵌套的
//  node.go：对 node 所存元素和 node 间关系的相关操作，节点内所存元素的增删、加载和落盘，访问孩子兄弟元素、拆分与合并的详细逻辑

// node represents an in-memory, deserialized page.
type node struct {
	bucket *Bucket // 其所在 bucket 的指针
	isLeaf bool    // 是否为叶子节点, branch node 或 leaf node

	// 调整、维持 B+ 树时使用
	unbalanced bool // 是否需要进行合并
	spilled    bool // 是否需要进行拆分和落盘

	key []byte // 所含第一个元素的 key

	pgid     pgid  // 对应的 page 的 id
	parent   *node // 父节点指针
	children nodes // 孩子节点指针（只包含加载到内存中的部分孩子）

	// 所存元素的元信息；对于分支节点branch node是 key+pgid 数组，对于叶子节点leaf node是 kv 数组
	inodes inodes // internal node，从 page 里读取的内容
}

// INFO: page 是磁盘上数据结构，需要读取 disk page(一个page默认4KB)为一个node，inode 就是 kvs，存储key-value数据
//  node.read page -> node
func (n *node) read(p *page) {
	// 初始化元信息
	n.pgid = p.id
	n.isLeaf = (p.flags & leafPageFlag) != 0
	n.inodes = make(inodes, int(p.count))

	// INFO: 加载所包含元素 inodes，因为B+tree里 branch node里存的是key，leaf node才会存key-value
	for i := 0; i < int(p.count); i++ {
		in := &n.inodes[i]
		if n.isLeaf {
			// 对于叶子节点leaf node是 kv 数组
			elem := p.leafPageElement(uint16(i))
			in.flags = elem.flags
			in.key = elem.key()
			in.value = elem.value()
		} else {
			// TODO: read branch page
			// 对于分支节点branch node是 key+pgid 数组
			elem := p.branchPageElement(uint16(i))
			in.pgid = elem.pgid
			in.key = elem.key()
		}
		if len(in.key) == 0 {
			klog.Fatalf(fmt.Sprintf("[node read]read from page to node err: zero-length inode key"))
		}
	}

	// 用第一个元素的 key 作为该 node 的 key，以便父节点以此作为索引进行查找和路由
	if len(n.inodes) > 0 {
		n.key = n.inodes[0].key
	} else {
		n.key = nil
	}
}

// INFO: 所有的数据新增都发生在叶子节点，如果新增数据后 B+ 树不平衡，之后会通过 node.spill 来进行拆分调整
//  该函数签名很有意思，同时有 oldKey 和 newKey ，开始时感觉很奇怪
//  1. 在叶子节点插入用户数据时，oldKey 等于 newKey，此时这两个参数是有冗余的。
//  2. 在 spill 阶段调整 B+ 树时，oldKey 可能不等于 newKey，此时是有用的，但从代码上来看，用处仍然很隐晦
//  为了避免在叶子节点最左侧插入一个很小的值时，引起祖先节点的 node.key 的链式更新，而将更新延迟到了最后 B+ 树调整阶段（spill 函数）进行统一处理
func (n *node) put(oldKey, newKey, value []byte, pgid pgid, flags uint32) {
	if pgid >= n.bucket.tx.meta.pgid { // pgid 不能大于 n.bucket.tx.meta.pgid
		panic(fmt.Sprintf("pgid (%d) above high water mark (%d)", pgid, n.bucket.tx.meta.pgid))
	} else if len(oldKey) <= 0 {
		panic("put: zero-length old key")
	} else if len(newKey) <= 0 {
		panic("put: zero-length new key")
	}

	// key还是升序的, 找到插入点
	index := sort.Search(len(n.inodes), func(i int) bool {
		return bytes.Compare(n.inodes[i].key, oldKey) != -1 // n.inodes[i].key < oldKey
	})

	// Add capacity and shift nodes if we don't have an exact match and need to insert.
	// 如果 key 是新增而非替换，则需为待插入节点腾空儿
	exact := len(n.inodes) > 0 && index < len(n.inodes) && bytes.Equal(n.inodes[index].key, oldKey)
	if !exact {
		n.inodes = append(n.inodes, inode{})
		copy(n.inodes[index+1:], n.inodes[index:]) // 腾空儿
	}

	// Loop over each item and write it to the page.
	off := unsafe.Sizeof(*p) + n.pageElementSize()*uintptr(len(n.inodes))
	for i, item := range n.inodes {
		_assert(len(item.key) > 0, "write: zero-length inode key")
		// INFO: allocate 出一个 ksize+vsize 空间出来, offset 指针已经后移
		//  buffer 指向数据的空间
		sz := len(item.key) + len(item.value)
		buffer := unsafeByteSlice(unsafe.Pointer(p), off, 0, sz)
		off += uintptr(sz)

		if n.isLeaf {
			elem := p.leafPageElement(uint16(i))
			elem.pos = uint32(uintptr(unsafe.Pointer(&buffer[0])) - uintptr(unsafe.Pointer(elem)))
			elem.flags = item.flags
			elem.ksize = uint32(len(item.key))
			elem.vsize = uint32(len(item.value))
		} else {
			// TODO: write branch node
		}

		// write data into leaf node
		l := copy(buffer, item.key)
		copy(buffer[l:], item.value)
	}
}

// spill writes the nodes to dirty pages and splits nodes as it goes.
// Returns an error if dirty pages cannot be allocated.
func (n *node) spill() error {
	var tx = n.bucket.tx
	if n.spilled {
		return nil
	}

	// Spill child nodes first. Child nodes can materialize sibling nodes in
	// the case of split-merge so we cannot use a range loop. We have to check
	// the children size on every loop iteration.
	sort.Sort(n.children)
	for i := 0; i < len(n.children); i++ {
		if err := n.children[i].spill(); err != nil {
			return err
		}
	}

	// We no longer need the child list because it's only used for spill tracking.
	n.children = nil
	// Split nodes into appropriate sizes. The first node will always be n.
	var nodes = n.split(uintptr(tx.db.pageSize))
	for _, node := range nodes {
		// Add node's page to the freelist if it's not new.
		if node.pgid > 0 {
			tx.db.freelist.free(tx.meta.txid, tx.page(node.pgid))
			node.pgid = 0
		}

		// Allocate contiguous space for the node.
		p, err := tx.allocate((node.size() + tx.db.pageSize - 1) / tx.db.pageSize)
		if err != nil {
			return err
		}

		// Write the node.
		if p.id >= tx.meta.pgid {
			panic(fmt.Sprintf("pgid (%d) above high water mark (%d)", p.id, tx.meta.pgid))
		}
		node.pgid = p.id
		node.write(p)
		node.spilled = true
	}
}

// split breaks up a node into multiple smaller nodes, if appropriate.
// This should only be called from the spill() function.
func (n *node) split(pageSize uintptr) []*node {
	var nodes []*node
	node := n
	for {
		// Split node into two.
		a, b := node.splitTwo(pageSize)
		nodes = append(nodes, a)

		// If we can't split then exit the loop.
		if b == nil {
			break
		}

		// Set node to b so it gets split on the next iteration.
		node = b
	}

	return nodes
}
func (n *node) splitTwo(pageSize uintptr) (*node, *node) {
	// Ignore the split if the page doesn't have at least enough nodes for
	// two pages or if the nodes can fit in a single page.
	if len(n.inodes) <= (minKeysPerPage*2) || n.sizeLessThan(pageSize) {
		return n, nil
	}

	// Determine the threshold before starting a new node.
	var fillPercent = n.bucket.FillPercent
	if fillPercent < minFillPercent {
		fillPercent = minFillPercent
	} else if fillPercent > maxFillPercent {
		fillPercent = maxFillPercent
	}
	threshold := int(float64(pageSize) * fillPercent)

	// Determine split position and sizes of the two pages.
	splitIndex, _ := n.splitIndex(threshold)

	// Split node into two separate nodes.
	// If there's no parent then we'll need to create one.
	if n.parent == nil {
		n.parent = &node{bucket: n.bucket, children: []*node{n}}
	}

	// Create a new node and add it to the parent.
	next := &node{bucket: n.bucket, isLeaf: n.isLeaf, parent: n.parent}
	n.parent.children = append(n.parent.children, next)

	// Split inodes across two nodes.
	next.inodes = n.inodes[splitIndex:]
	n.inodes = n.inodes[:splitIndex]

	// Update the statistics.
	n.bucket.tx.stats.Split++

	return n, next
}

// splitIndex finds the position where a page will fill a given threshold.
// It returns the index as well as the size of the first page.
// This is only be called from split().
func (n *node) splitIndex(threshold int) (index, sz uintptr) {
	sz = pageHeaderSize

	// Loop until we only have the minimum number of keys required for the second page.
	for i := 0; i < len(n.inodes)-minKeysPerPage; i++ {
		index = uintptr(i)
		inode := n.inodes[i]
		elsize := n.pageElementSize() + uintptr(len(inode.key)) + uintptr(len(inode.value))
		// If we have at least the minimum number of keys and adding another
		// node would put us over the threshold then exit and return.
		if i >= minKeysPerPage && sz+elsize > uintptr(threshold) {
			break
		}

		// Add the element size to the total size.
		sz += elsize
	}

	return
}

// write writes the items onto one or more pages.
func (n *node) write(p *page) {
	// Initialize page.
	if n.isLeaf {
		p.flags |= leafPageFlag
	} else {
		p.flags |= branchPageFlag
	}

	if len(n.inodes) >= 0xFFFF {
		panic(fmt.Sprintf("inode overflow: %d (pgid=%d)", len(n.inodes), p.id))
	}
	p.count = uint16(len(n.inodes))

	// Stop here if there are no items to write.
	if p.count == 0 {
		return
	}

	// Loop over each item and write it to the page.
	b := (*[maxAllocSize]byte)(unsafe.Pointer(&p.ptr))[n.pageElementSize()*uintptr(len(n.inodes)):]
	for i, item := range n.inodes {
		_assert(len(item.key) > 0, "write: zero-length inode key")

		// Write the page element.
		if n.isLeaf {
			elem := p.leafPageElement(uint16(i))
			elem.pos = uint32(uintptr(unsafe.Pointer(&b[0])) - uintptr(unsafe.Pointer(elem)))
			elem.flags = item.flags
			elem.ksize = uint32(len(item.key))
			elem.vsize = uint32(len(item.value))
		} else {

		}

		// If the length of key+value is larger than the max allocation size
		// then we need to reallocate the byte array pointer.
		//
		// See: https://github.com/boltdb/bolt/pull/335
		klen, vlen := len(item.key), len(item.value)
		if len(b) < klen+vlen {
			b = (*[maxAllocSize]byte)(unsafe.Pointer(&b[0]))[:]
		}

		// Write data for the element to the end of the page.
		copy(b[0:], item.key)
		b = b[klen:]
		copy(b[0:], item.value)
		b = b[vlen:]
	}
}

// pageElementSize returns the size of each page element based on the type of node.
func (n *node) pageElementSize() uintptr {
	if n.isLeaf {
		return leafPageElementSize
	}
	return branchPageElementSize
}

// sizeLessThan returns true if the node is less than a given size.
// This is an optimization to avoid calculating a large node when we only need
// to know if it fits inside a certain page size.
func (n *node) sizeLessThan(v uintptr) bool {
	sz, elsz := pageHeaderSize, n.pageElementSize()
	for i := 0; i < len(n.inodes); i++ {
		item := &n.inodes[i]
		sz += elsz + uintptr(len(item.key)) + uintptr(len(item.value))
		if sz >= v {
			return false
		}
	}
	return true
}

// size returns the size of the node after serialization.
func (n *node) size() int {
	sz, elsz := pageHeaderSize, n.pageElementSize()
	for i := 0; i < len(n.inodes); i++ {
		item := &n.inodes[i]
		sz += elsz + uintptr(len(item.key)) + uintptr(len(item.value))
	}
	return int(sz)
}
