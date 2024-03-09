package buffer

import "unsafe"

//go:noescape

// Zero the n bytes starting at p.
//
// REQUIRES: the region does not contain any Go pointers.
//
//go:linkname jacobsa_fuse_memclr runtime.memclrNoHeapPointers
func jacobsa_fuse_memclr(p unsafe.Pointer, n uintptr)

//go:noescape

// Copy from src to dst, allowing overlap.
//
//go:linkname jacobsa_fuse_memmove runtime.memmove
func jacobsa_fuse_memmove(dst unsafe.Pointer, src unsafe.Pointer, n uintptr)
