package freelist

import "unsafe"

// A freelist for arbitrary pointers. Not safe for concurrent access.
type Freelist struct {
	list []unsafe.Pointer
}

// Get an element from the freelist, returning nil if empty.
func (fl *Freelist) Get() unsafe.Pointer {
	l := len(fl.list)
	if l == 0 {
		return nil
	}

	p := fl.list[l-1]
	fl.list = fl.list[:l-1]

	return p
}

// Contribute an element back to the freelist.
func (fl *Freelist) Put(p unsafe.Pointer) {
	fl.list = append(fl.list, p)
}
