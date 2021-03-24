package freelist

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
	freelistType   FreelistType                // freelist type
	ids            []pgid                      // all free and available free page ids.
	allocs         map[pgid]txid               // mapping of txid that allocated a pgid.
	pending        map[txid]*txPending         // mapping of soon-to-be free page ids by tx.
	cache          map[pgid]bool               // fast lookup of all free and pending page ids.
	freemaps       map[uint64]pidSet           // key is the size of continuous pages(span), value is a set which contains the starting pgids of same size
	forwardMap     map[pgid]uint64             // key is start pgid, value is its span size
	backwardMap    map[pgid]uint64             // key is end pgid, value is its span size
	allocate       func(txid txid, n int) pgid // the freelist allocate func
	free_count     func() int                  // the function which gives you free page number
	mergeSpans     func(ids pgids)             // the mergeSpan func
	getFreePageIDs func() []pgid               // get free pgids func
	readIDs        func(pgids []pgid)          // readIDs func reads list of pages and init the freelist
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
func newFreelist(freelistType FreelistType) *freelist {
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
}
