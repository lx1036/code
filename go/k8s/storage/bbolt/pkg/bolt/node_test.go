package bolt

import (
	"testing"
	"unsafe"
)

func TestNodePut(test *testing.T) {
	n := &node{inodes: make(inodes, 0), bucket: &Bucket{tx: &Tx{meta: &meta{pgid: 1}}}}
	n.put([]byte("baz"), []byte("baz"), []byte("2"), 0, 0)
	n.put([]byte("foo"), []byte("foo"), []byte("0"), 0, 0)
	n.put([]byte("bar"), []byte("bar"), []byte("1"), 0, 0)
	n.put([]byte("foo"), []byte("foo"), []byte("3"), 0, leafPageFlag)
	// bar=>1, barz=>2, foo=>3

	if len(n.inodes) != 3 {
		test.Fatalf("exp=3; got=%d", len(n.inodes))
	}
	if k, v := n.inodes[0].key, n.inodes[0].value; string(k) != "bar" || string(v) != "1" {
		test.Fatalf("exp=<bar,1>; got=<%s,%s>", k, v)
	}
	if k, v := n.inodes[1].key, n.inodes[1].value; string(k) != "baz" || string(v) != "2" {
		test.Fatalf("exp=<baz,2>; got=<%s,%s>", k, v)
	}
	if k, v := n.inodes[2].key, n.inodes[2].value; string(k) != "foo" || string(v) != "3" {
		test.Fatalf("exp=<foo,3>; got=<%s,%s>", k, v)
	}
	if n.inodes[2].flags != uint32(leafPageFlag) {
		test.Fatalf("not a leaf: %d", n.inodes[2].flags)
	}
}

func TestNodeReadLeftPage(test *testing.T) {
	// Create a page.
	var buf [4096]byte
	page := (*page)(unsafe.Pointer(&buf[0]))
	page.flags = leafPageFlag
	page.count = 2

	// Insert 2 elements at the beginning. sizeof(leafPageElement) == 16
	nodes := (*[3]leafPageElement)(unsafe.Pointer(&page.ptr))
	nodes[0] = leafPageElement{flags: 0, pos: 32, ksize: 3, vsize: 4}  // pos = sizeof(leafPageElement) * 2
	nodes[1] = leafPageElement{flags: 0, pos: 23, ksize: 10, vsize: 3} // pos = sizeof(leafPageElement) + 3 + 4
	// Write data for the nodes at the end.
	data := (*[4096]byte)(unsafe.Pointer(&nodes[2]))
	copy(data[:], []byte("barfooz"))
	copy(data[7:], []byte("helloworldbye"))

	// Deserialize page into a leaf.
	// node读取page
	n := &node{}
	n.read(page)
	// Check that there are two inodes with correct data.
	if !n.isLeaf {
		test.Fatal("expected leaf")
	}
	if len(n.inodes) != 2 {
		test.Fatalf("exp=2; got=%d", len(n.inodes))
	}
	if k, v := n.inodes[0].key, n.inodes[0].value; string(k) != "bar" || string(v) != "fooz" {
		test.Fatalf("exp=<bar,fooz>; got=<%s,%s>", k, v)
	}
	if k, v := n.inodes[1].key, n.inodes[1].value; string(k) != "helloworld" || string(v) != "bye" {
		test.Fatalf("exp=<helloworld,bye>; got=<%s,%s>", k, v)
	}
}

// Ensure that a node can serialize into a leaf page.
func TestNodeWriteLeftPage(test *testing.T) {
	// Create a node.
	n := &node{isLeaf: true, inodes: make(inodes, 0), bucket: &Bucket{tx: &Tx{db: &DB{}, meta: &meta{pgid: 1}}}}
	n.put([]byte("susy"), []byte("susy"), []byte("que"), 0, 0)
	n.put([]byte("ricki"), []byte("ricki"), []byte("lake"), 0, 0)
	n.put([]byte("john"), []byte("john"), []byte("johnson"), 0, 0)

	// Write it to a page.
	var buf [4096]byte
	p := (*page)(unsafe.Pointer(&buf[0]))
	n.write(p)

	// Read the page back in.
	n2 := &node{}
	n2.read(p)

	// Check that the two pages are the same.
	if len(n2.inodes) != 3 {
		test.Fatalf("exp=3; got=%d", len(n2.inodes))
	}
	if k, v := n2.inodes[0].key, n2.inodes[0].value; string(k) != "john" || string(v) != "johnson" {
		test.Fatalf("exp=<john,johnson>; got=<%s,%s>", k, v)
	}
	if k, v := n2.inodes[1].key, n2.inodes[1].value; string(k) != "ricki" || string(v) != "lake" {
		test.Fatalf("exp=<ricki,lake>; got=<%s,%s>", k, v)
	}
	if k, v := n2.inodes[2].key, n2.inodes[2].value; string(k) != "susy" || string(v) != "que" {
		test.Fatalf("exp=<susy,que>; got=<%s,%s>", k, v)
	}
}

func TestSplit(test *testing.T) {
	// Create a node.
	n := &node{inodes: make(inodes, 0), bucket: &Bucket{tx: &Tx{db: &DB{}, meta: &meta{pgid: 1}}}}
	n.put([]byte("00000001"), []byte("00000001"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000002"), []byte("00000002"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000003"), []byte("00000003"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000004"), []byte("00000004"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000005"), []byte("00000005"), []byte("0123456701234567"), 0, 0)
	// Split between 2 & 3.
	n.split(100)
	var parent = n.parent
	if len(parent.children) != 2 {
		test.Fatalf("exp=2; got=%d", len(parent.children))
	}
	if len(parent.children[0].inodes) != 2 {
		test.Fatalf("exp=2; got=%d", len(parent.children[0].inodes))
	}
	if len(parent.children[1].inodes) != 3 {
		test.Fatalf("exp=3; got=%d", len(parent.children[1].inodes))
	}

	// split min keys
	n2 := &node{inodes: make(inodes, 0), bucket: &Bucket{tx: &Tx{db: &DB{}, meta: &meta{pgid: 1}}}}
	n2.put([]byte("00000001"), []byte("00000001"), []byte("0123456701234567"), 0, 0)
	n2.put([]byte("00000002"), []byte("00000002"), []byte("0123456701234567"), 0, 0)
	// Split.
	n2.split(20)
	if n2.parent != nil {
		test.Fatalf("expected nil parent")
	}

	// split single page
	n3 := &node{inodes: make(inodes, 0), bucket: &Bucket{tx: &Tx{db: &DB{}, meta: &meta{pgid: 1}}}}
	n3.put([]byte("00000001"), []byte("00000001"), []byte("0123456701234567"), 0, 0)
	n3.put([]byte("00000002"), []byte("00000002"), []byte("0123456701234567"), 0, 0)
	n3.put([]byte("00000003"), []byte("00000003"), []byte("0123456701234567"), 0, 0)
	n3.put([]byte("00000004"), []byte("00000004"), []byte("0123456701234567"), 0, 0)
	n3.put([]byte("00000005"), []byte("00000005"), []byte("0123456701234567"), 0, 0)
	n3.split(4096) // 4k
	if n3.parent != nil {
		test.Fatalf("expected nil parent")
	}
}
