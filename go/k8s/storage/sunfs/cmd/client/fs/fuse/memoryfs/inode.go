package main

import (
	"time"

	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
)

// Common attributes for files and directories.
//
// External synchronization is required.
type inode struct {

	// For directories, entries describing the children of the directory. Unused
	// entries are of type DT_Unknown.
	//
	// This array can never be shortened, nor can its elements be moved, because
	// we use its indices for Dirent.Offset, which is exposed to the user who
	// might be calling readdir in a loop while concurrently modifying the
	// directory. Unused entries can, however, be reused.
	//
	// INVARIANT: If !isDir(), len(entries) == 0
	// INVARIANT: For each i, entries[i].Offset == i+1
	// INVARIANT: Contains no duplicate names in used entries.
	entries []fuseutil.Dirent

	// The current attributes of this inode.
	//
	// INVARIANT: attrs.Mode &^ (os.ModePerm|os.ModeDir|os.ModeSymlink) == 0
	// INVARIANT: !(isDir() && isSymlink())
	// INVARIANT: attrs.Size == len(contents)
	attrs fuseops.InodeAttributes

	// extended attributes and values
	xattrs map[string][]byte
}

// Return the index of the child within in.entries, if it exists.
//
// REQUIRES: in.isDir()
func (in *inode) findChild(name string) (i int, ok bool) {
	if !in.isDir() {
		panic("findChild called on non-directory.")
	}

	var e fuseutil.Dirent
	for i, e = range in.entries {
		if e.Name == name {
			return i, true
		}
	}

	return 0, false
}

func (in *inode) isDir() bool {
	return in.attrs.Mode.IsDir()
}

// Serve a ReadDir request.
//
// REQUIRES: in.isDir()
func (in *inode) ReadDir(dst []byte, offset int) int {
	if !in.isDir() {
		panic("ReadDir called on non-directory.")
	}

	var n int
	for i := offset; i < len(in.entries); i++ {
		entry := in.entries[i]
		// Skip unused entries.
		if entry.Type == fuseutil.DT_Unknown {
			continue
		}

		tmp := fuseutil.WriteDirent(dst[n:], in.entries[i])
		if tmp == 0 {
			break
		}

		n += tmp
	}

	return n
}

// Find an entry for the given child name and return its inode ID.
//
// REQUIRES: in.isDir()
func (in *inode) LookUpChild(name string) (id fuseops.InodeID, typ fuseutil.DirentType, ok bool) {
	index, ok := in.findChild(name)
	if ok {
		id = in.entries[index].Inode
		typ = in.entries[index].Type
	}

	return id, typ, ok
}

// Add an entry for a child.
//
// REQUIRES: in.isDir()
// REQUIRES: dt != fuseutil.DT_Unknown
func (in *inode) AddChild(id fuseops.InodeID, name string, dt fuseutil.DirentType) {
	var index int

	// Update the modification time.
	in.attrs.Mtime = time.Now()

	e := fuseutil.Dirent{
		Inode: id,
		Name:  name,
		Type:  dt,
	}

	// No matter where we place the entry, make sure it has the correct Offset
	// field.
	defer func() {
		in.entries[index].Offset = fuseops.DirOffset(index + 1)
	}()

	// INFO: 如果有 Unknown entry，则用这个 entry 补位
	for index = range in.entries {
		if in.entries[index].Type == fuseutil.DT_Unknown {
			in.entries[index] = e
			return
		}
	}

	// Append it to the end.
	index = len(in.entries)
	in.entries = append(in.entries, e)
}

// Create a new inode with the supplied attributes, which need not contain
// time-related information (the inode object will take care of that).
func newInode(attrs fuseops.InodeAttributes) *inode {
	// Update time info.
	now := time.Now()
	attrs.Mtime = now
	attrs.Crtime = now

	// Create the object.
	return &inode{
		attrs:  attrs,
		xattrs: make(map[string][]byte),
	}
}
