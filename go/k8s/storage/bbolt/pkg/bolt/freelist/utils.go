package freelist

import "unsafe"

// maxAllocSize is the size used when creating array pointers.
const maxAllocSize = 0x7FFFFFFF

func unsafeAdd(base unsafe.Pointer, offset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(base) + offset)
}

func unsafeByteSlice(base unsafe.Pointer, offset uintptr, i, j int) []byte {
	// See: https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices
	//
	// This memory is not allocated from C, but it is unmanaged by Go's
	// garbage collector and should behave similarly, and the compiler
	// should produce similar code.  Note that this conversion allows a
	// subslice to begin after the base address, with an optional offset,
	// while the URL above does not cover this case and only slices from
	// index 0.  However, the wiki never says that the address must be to
	// the beginning of a C allocation (or even that malloc was used at
	// all), so this is believed to be correct.
	return (*[maxAllocSize]byte)(unsafeAdd(base, offset))[i:j:j]
}
