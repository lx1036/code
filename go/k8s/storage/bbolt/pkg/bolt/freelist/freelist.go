package freelist

import (
	"fmt"
	"sort"
	"unsafe"
)

//******************************
// https://github.com/boltdb/bolt
//******************************

const pgidNoFreelist pgid = 0xffffffffffffffff

// FreelistType is the type of the freelist backend
type FreelistType string

const (
	// FreelistArrayType indicates backend freelist type is array
	FreelistArrayType = FreelistType("array")
	// FreelistMapType indicates backend freelist type is hashmap
	FreelistMapType = FreelistType("hashmap")
)

// txid represents the internal transaction identifier.
type txid uint64

// pidSet holds the set of starting pgids which have the same span size
type pidSet map[pgid]struct{}

// txPending holds a list of pgids and corresponding allocation txns
// that are pending to be freed.
type txPending struct {
	ids              []pgid
	alloctx          []txid // txids allocating the ids
	lastReleaseBegin txid   // beginning txid of last matching releaseRange
}

// freelist表示可以未被分配/已经被释放的[]pages
// 空闲列表
type freelist struct {
	//freelistType   FreelistType                // freelist type
	ids []pgid // 已经被release的page ids
	//allocs         map[pgid]txid               // mapping of txid that allocated a pgid.
	//pending        map[txid]*txPending         // mapping of soon-to-be free page ids by tx.
	pending map[txid][]pgid // 存放当前事务要被free的 page ids
	cache   map[pgid]bool   // 缓存判断当前page id是否需要被free，可以快速查找
	//freemaps       map[uint64]pidSet           // key is the size of continuous pages(span), value is a set which contains the starting pgids of same size
	//forwardMap     map[pgid]uint64             // key is start pgid, value is its span size
	//backwardMap    map[pgid]uint64             // key is end pgid, value is its span size
	//allocate       func(txid txid, n int) pgid // the freelist allocate func
	//free_count     func() int                  // the function which gives you free page number
	//mergeSpans     func(ids pgids)             // the mergeSpan func
	//getFreePageIDs func() []pgid               // get free pgids func
	//readIDs        func(pgids []pgid)          // readIDs func reads list of pages and init the freelist
}

func newFreelist() *freelist {
	return &freelist{
		pending: make(map[txid][]pgid),
		cache:   make(map[pgid]bool),
	}
}

// free releases a page and its overflow for a given transaction id.
// If the page is already free then a panic will occur.
// free只是标记要被release的pages
func (f *freelist) free(txid txid, p *page) {
	if p.id <= 1 {
		panic(fmt.Sprintf("cannot free page 0 or 1: %d", p.id))
	}
	// Free page and all its overflow pages.
	//ids := f.pending[txid]
	for id := p.id; id <= p.id+pgid(p.overflow); id++ {
		// Verify that page is not already free.
		if f.cache[id] {
			panic(fmt.Sprintf("page %d already freed", id))
		}

		// Add to the freelist and cache.
		f.pending[txid] = append(f.pending[txid], id)
		f.cache[id] = true
	}
}

// release moves all page ids for a transaction id (or older) to the freelist.
func (f *freelist) release(txid txid) {
	m := make(pgids, 0)
	for tid, ids := range f.pending {
		if tid <= txid { // 删除过往所有tid
			// Move transaction's pending pages to the available freelist.
			// Don't remove from the cache since the page is still free.
			m = append(m, ids...)
			// 真正删除tid
			delete(f.pending, tid)
		}
	}
	sort.Sort(m)
	f.ids = pgids(f.ids).merge(m) // merge之后还是排序好的
}

// 分配连续的n个page，这里连续意思是：pgid,pgid+1,pgid+2
// 比如{1,2,3,5,6,7,8,10}，allocate(4)就是分配了{5,6,7,8}出去，剩下{1,2,3}
func (f *freelist) allocate(n int) pgid {
	if len(f.ids) == 0 {
		return 0
	}

	var initial, previd pgid
	for i, id := range f.ids {
		if id <= 1 {
			panic(fmt.Sprintf("invalid page allocation: %d", id))
		}

		if previd == 0 || id-previd != 1 {
			initial = id // initial是分配段初始值
		}

		tmp := id - initial + 1
		if tmp == pgid(n) {
			if i+1 == n {
				f.ids = f.ids[i+1:] // 3,4,5被分配出去了
			} else {
				// {7,9,12,13,18} => 拿{18}去覆盖{12,13,18} => {7,9,18}
				copy(f.ids[i+1-n:], f.ids[i+1:])
				f.ids = f.ids[:len(f.ids)-n]
			}

			// Remove from the free cache.
			for i := pgid(0); i < pgid(n); i++ {
				delete(f.cache, initial+i)
			}

			return initial
		}

		previd = id
	}

	return 0
}

// read initializes the freelist from a freelist page.
func (f *freelist) read(p *page) {
	// If the page.count is at the max uint16 value (64k) then it's considered
	// an overflow and the size of the freelist is stored as the first element.
	idx, count := 0, int(p.count)
	if count == 0xFFFF {
		idx = 1
		count = int(((*[maxAllocSize]pgid)(unsafe.Pointer(&p.ptr)))[0])
	}

}

// hashmapAllocate serves the same purpose as arrayAllocate, but use hashmap as backend
func (f *freelist) hashmapAllocate(txid txid, n int) pgid {
	return 0
}

// arrayAllocate returns the starting page id of a contiguous list of pages of a given size.
// If a contiguous block cannot be found then 0 is returned.
func (f *freelist) arrayAllocate(txid txid, n int) pgid {
	return 0
}

// hashmapFreeCount returns count of free pages(hashmap version)
func (f *freelist) hashmapFreeCount() int {
	return 0
}

// arrayFreeCount returns count of free pages(array version)
func (f *freelist) arrayFreeCount() int {
	return 0
}

// hashmapMergeSpans try to merge list of pages(represented by pgids) with existing spans
func (f *freelist) hashmapMergeSpans(ids pgids) {

}

// arrayMergeSpans try to merge list of pages(represented by pgids) with existing spans but using array
func (f *freelist) arrayMergeSpans(ids pgids) {

}

// hashmapGetFreePageIDs returns the sorted free page ids
func (f *freelist) hashmapGetFreePageIDs() []pgid {
	return nil
}
func (f *freelist) arrayGetFreePageIDs() []pgid {
	return f.ids
}

// hashmapReadIDs reads pgids as input an initial the freelist(hashmap version)
func (f *freelist) hashmapReadIDs(pgids []pgid) {

}

// arrayReadIDs initializes the freelist from a given list of ids.
func (f *freelist) arrayReadIDs(ids []pgid) {

}

// newFreelist returns an empty, initialized freelist.
/*func newFreelist(freelistType FreelistType) *freelist {
	f := &freelist{
		freelistType: freelistType,
		allocs:       make(map[pgid]txid),
		pending:      make(map[txid]*txPending),
		cache:        make(map[pgid]bool),
		freemaps:     make(map[uint64]pidSet),
		forwardMap:   make(map[pgid]uint64),
		backwardMap:  make(map[pgid]uint64),
	}

	// 先不着急实现hashmap type
	if freelistType == FreelistMapType {
		f.allocate = f.hashmapAllocate
		f.free_count = f.hashmapFreeCount
		f.mergeSpans = f.hashmapMergeSpans
		f.getFreePageIDs = f.hashmapGetFreePageIDs
		f.readIDs = f.hashmapReadIDs
	} else {
		f.allocate = f.arrayAllocate
		f.free_count = f.arrayFreeCount
		f.mergeSpans = f.arrayMergeSpans
		f.getFreePageIDs = f.arrayGetFreePageIDs
		f.readIDs = f.arrayReadIDs
	}

	return f
}*/
