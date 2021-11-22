package bolt

import (
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"unsafe"
)

// INFO: boltdb page 概念 https://time.geekbang.org/column/article/342527

const branchPageElementSize = unsafe.Sizeof(branchPageElement{})
const leafPageElementSize = unsafe.Sizeof(leafPageElement{})

const minKeysPerPage = 2
const pageHeaderSize = unsafe.Offsetof(((*page)(nil)).ptr)

const (
	bucketLeafFlag = 0x01
)

type pgid uint64
type pgids []pgid

func (s pgids) Len() int           { return len(s) }
func (s pgids) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s pgids) Less(i, j int) bool { return s[i] < s[j] }

// merge returns the sorted union of a and b.
func (s pgids) merge(b pgids) pgids {
	// Return the opposite slice if one is nil.
	if len(s) == 0 {
		return b
	}
	if len(b) == 0 {
		return s
	}

	merged := make(pgids, len(s)+len(b))
	mergepgids(merged, s, b)
	return merged
}

func mergepgids(dst, a, b pgids) {
	if len(dst) < len(a)+len(b) {
		panic(fmt.Errorf("mergepgids bad len %d < %d + %d", len(dst), len(a), len(b)))
	}
	// Copy in the opposite slice if one is nil.
	if len(a) == 0 {
		copy(dst, b)
		return
	}
	if len(b) == 0 {
		copy(dst, a)
		return
	}

	// ?????????
	// 这块不知道用了什么算法merge了两个排序数组，leetcode上找下

	// 首元素最小当lead
	lead, follow := a, b
	if b[0] < a[0] {
		lead, follow = b, a
	}

	// Merged will hold all elements from both lists.
	merged := dst[:0] // []

	// Continue while there are elements in the lead.
	for len(lead) > 0 {
		// Merge largest prefix of lead that is ahead of follow[0].
		n := sort.Search(len(lead), func(i int) bool { return lead[i] > follow[0] })
		merged = append(merged, lead[:n]...)
		if n >= len(lead) {
			break
		}

		// Swap lead and follow.
		lead, follow = follow, lead[n:]
	}

	// Append what's left in follow.
	_ = append(merged, follow...)
}

const (
	branchPageFlag   = 0x01 // branch page
	leafPageFlag     = 0x02 // leaf page
	metaPageFlag     = 0x04 // meta page
	freelistPageFlag = 0x10 // freelist page
)

// page 是操作系统页大小，读写数据最小原子单位
type page struct {
	id       pgid    // 页ID
	flags    uint16  // 页类型，这块内容标识：可以为元数据、空闲列表、树枝、叶子 这四种中的一种
	count    uint16  // 数量，存储数据的数量
	overflow uint32  // 溢出页数量，溢出的页数量
	ptr      uintptr // 页数据起始位置，内存中存储数据的指针，没有落盘
}

// typ returns a human readable page type string used for debugging.
func (p *page) typ() string {
	if (p.flags & branchPageFlag) != 0 {
		return "branch"
	} else if (p.flags & leafPageFlag) != 0 {
		return "leaf"
	} else if (p.flags & metaPageFlag) != 0 {
		return "meta"
	} else if (p.flags & freelistPageFlag) != 0 {
		return "freelist"
	}
	return fmt.Sprintf("unknown<%02x>", p.flags)
}

// dump writes n bytes of the page to STDERR as hex output.
func (p *page) hexdump(n int) {
	buf := unsafeByteSlice(unsafe.Pointer(p), 0, 0, n)
	fmt.Fprintf(os.Stderr, "%x\n", buf)
}

// leafPageElement retrieves the leaf node by index
func (p *page) leafPageElement(index uint16) *leafPageElement {
	return (*leafPageElement)(unsafeIndex(unsafe.Pointer(p), unsafe.Sizeof(*p),
		leafPageElementSize, int(index)))
}

// branchPageElement retrieves the branch node by index
func (p *page) branchPageElement(index uint16) *branchPageElement {
	return &((*[0x7FFFFFF]branchPageElement)(unsafe.Pointer(&p.ptr)))[index]
}

// meta returns a pointer to the metadata section of the page.
func (p *page) meta() *meta {
	return (*meta)(unsafe.Pointer(&p.ptr))
}

type pages []*page

func (s pages) Len() int           { return len(s) }
func (s pages) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s pages) Less(i, j int) bool { return s[i].id < s[j].id }

// INFO: 第 0、1 页它是固定存储 db 元数据的页 (meta page), `bbolt dump db 0`
/*
magic version pageSize flags root freelist pgid txid checksum
0001000 0100 0000 0000 0000 0400 0000 0000 0000
0001010 edda 0ced 0200 0000 0010 0000 0000 0000
0001020 0200 0000 0000 0000 0000 0000 0000 0000
0001030 0300 0000 0000 0000 0600 0000 0000 0000
0001040 0d00 0000 0000 0000 414e 805d ac14 838c
0001050 0000 0000 0000 0000 0000 0000 0000 0000
0001060 *
0001ff0 0000 0000 0000 0000 0000 0000 0000 0000
*/

type meta struct {
	magic    uint32 // 文件标识
	version  uint32 // 版本号
	pageSize uint32 // 页大小
	flags    uint32 // 页类型
	root     bucket // 根bucket
	freelist pgid   // freelist页面id
	pgid     pgid   // 总的页面数量
	txid     txid   // 上一次写事务id
	checksum uint64 // 校验码
}

// validate checks the marker bytes and version of the meta page to ensure it matches this binary.
func (m *meta) validate() error {
	if m.magic != magic {
		return ErrInvalid
	} else if m.version != version {
		return ErrVersionMismatch
	} else if m.checksum != 0 && m.checksum != m.sum64() {
		return ErrChecksum
	}
	return nil
}
func (m *meta) sum64() uint64 {
	var h = fnv.New64a()
	_, _ = h.Write((*[unsafe.Offsetof(meta{}.checksum)]byte)(unsafe.Pointer(m))[:])
	return h.Sum64()
}

// leafPageElement represents a node on a leaf page.
type leafPageElement struct {
	flags uint32
	pos   uint32
	ksize uint32
	vsize uint32
}

// key returns a byte slice of the node key.
func (n *leafPageElement) key() []byte {
	i := int(n.pos)
	j := i + int(n.ksize)
	return unsafeByteSlice(unsafe.Pointer(n), 0, i, j)
}

// value returns a byte slice of the node value.
func (n *leafPageElement) value() []byte {
	i := int(n.pos) + int(n.ksize)
	j := i + int(n.vsize)
	return unsafeByteSlice(unsafe.Pointer(n), 0, i, j)
}

// branchPageElement represents a node on a branch page.
type branchPageElement struct {
	pos   uint32
	ksize uint32
	pgid  pgid
}

// key returns a byte slice of the node key.
func (n *branchPageElement) key() []byte {
	return unsafeByteSlice(unsafe.Pointer(n), 0, int(n.pos), int(n.pos)+int(n.ksize))
}
