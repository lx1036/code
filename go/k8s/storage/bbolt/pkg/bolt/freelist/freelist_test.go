package freelist

import (
	"reflect"
	"testing"
	"unsafe"
)

// Ensure that a page is added to a transaction's freelist.
func TestFreelistFree(test *testing.T) {
	list := newFreelist()
	list.free(100, &page{id: 12})
	if !reflect.DeepEqual([]pgid{12}, list.pending[100]) {
		test.Fatalf("exp=%v; got=%v", []pgid{12}, list.pending[100])
	}
}

// Ensure that a page and its overflow is added to a transaction's freelist.
func TestFreelist_free_overflow(t *testing.T) {
	f := newFreelist()
	f.free(100, &page{id: 12, overflow: 3})
	if exp := []pgid{12, 13, 14, 15}; !reflect.DeepEqual(exp, f.pending[100]) {
		t.Fatalf("exp=%v; got=%v", exp, f.pending[100])
	}
}

// Ensure that a transaction's free pages can be released.
func TestFreelist_release(t *testing.T) {
	f := newFreelist()
	f.free(100, &page{id: 12, overflow: 1}) // 12,12+1
	f.free(100, &page{id: 9})
	f.free(102, &page{id: 39})
	f.release(100)
	f.release(101) // 执行两次，9,12,13 会从pending中删除，进入f.ids中
	if exp := []pgid{9, 12, 13}; !reflect.DeepEqual(exp, f.ids) {
		t.Fatalf("exp=%v; got=%v", exp, f.ids)
	}

	f.release(102) // 9,12,13,39 真正删除，进入f.ids中
	if exp := []pgid{9, 12, 13, 39}; !reflect.DeepEqual(exp, f.ids) {
		t.Fatalf("exp=%v; got=%v", exp, f.ids)
	}
}

// allocate用来分配page id连续的page出去
func TestFreelist_allocate(t *testing.T) {
	f := &freelist{ids: []pgid{3, 4, 5, 6, 7, 9, 12, 13, 18}}
	if id := int(f.allocate(3)); id != 3 {
		t.Fatalf("exp=3; got=%v", id)
	}
	if id := int(f.allocate(1)); id != 6 {
		t.Fatalf("exp=6; got=%v", id)
	}
	if id := int(f.allocate(3)); id != 0 {
		t.Fatalf("exp=0; got=%v", id)
	}
	if id := int(f.allocate(2)); id != 12 {
		t.Fatalf("exp=12; got=%v", id)
	}
	if id := int(f.allocate(1)); id != 7 {
		t.Fatalf("exp=7; got=%v", id)
	}
	if id := int(f.allocate(0)); id != 0 { // 分配0个出去，会依次遍历所有page ids
		t.Fatalf("exp=0; got=%v", id)
	}
	if id := int(f.allocate(0)); id != 0 {
		t.Fatalf("exp=0; got=%v", id)
	}

	if exp := []pgid{9, 18}; !reflect.DeepEqual(exp, f.ids) {
		t.Fatalf("exp=%v; got=%v", exp, f.ids)
	}

	if id := int(f.allocate(1)); id != 9 {
		t.Fatalf("exp=9; got=%v", id)
	}
	if id := int(f.allocate(1)); id != 18 {
		t.Fatalf("exp=18; got=%v", id)
	}
	if id := int(f.allocate(1)); id != 0 {
		t.Fatalf("exp=0; got=%v", id)
	}
	if exp := []pgid{}; !reflect.DeepEqual(exp, f.ids) {
		t.Fatalf("exp=%v; got=%v", exp, f.ids)
	}
}

// Ensure that a freelist can deserialize from a freelist page.
func TestFreelist_read(t *testing.T) {
	// Create a page.
	var buf [4096]byte
	page := (*page)(unsafe.Pointer(&buf[0]))
	page.flags = freelistPageFlag
	page.count = 2

	// Insert 2 page ids.
	ids := (*[3]pgid)(unsafe.Pointer(&page.ptr))
	ids[0] = 23
	ids[1] = 50

	// Deserialize page into a freelist.
	f := newFreelist()
	f.read(page)

	// Ensure that there are two page ids in the freelist.
	if exp := []pgid{23, 50}; !reflect.DeepEqual(exp, f.ids) {
		t.Fatalf("exp=%v; got=%v", exp, f.ids)
	}
}
