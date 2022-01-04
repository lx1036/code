package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
)

// Common attributes for files and directories.
//
// External synchronization is required.
type inode struct {

	// For files, the current contents of the file.
	//
	// INVARIANT: If !isFile(), len(contents) == 0
	contents []byte

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

	// For symlinks, the target of the symlink.
	//
	// INVARIANT: If !isSymlink(), len(target) == 0
	target string
}

func newInode(attrs fuseops.InodeAttributes) *inode {
	now := time.Now()
	attrs.Mtime = now
	attrs.Crtime = now

	// Create the object.
	return &inode{
		attrs:  attrs,
		xattrs: make(map[string][]byte),
	}
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

// Remove an entry for a child.
//
// REQUIRES: in.isDir()
// REQUIRES: An entry for the given name exists.
func (in *inode) RemoveChild(name string) {
	// Update the modification time.
	in.attrs.Mtime = time.Now()

	// Find the entry.
	i, ok := in.findChild(name)
	if !ok {
		panic(fmt.Sprintf("Unknown child: %s", name))
	}

	// Mark it as unused.
	in.entries[i] = fuseutil.Dirent{
		Type:   fuseutil.DT_Unknown,
		Offset: fuseops.DirOffset(i + 1),
	}
}

// Return the number of children of the directory.
//
// REQUIRES: in.isDir()
func (in *inode) Len() int {
	var n int
	for _, e := range in.entries {
		if e.Type != fuseutil.DT_Unknown {
			n++
		}
	}

	return n
}

// Write to the file's contents. See documentation for ioutil.WriterAt.
//
// REQUIRES: in.isFile()
func (in *inode) WriteAt(p []byte, off int64) (int, error) {
	if !in.isFile() {
		panic("WriteAt called on non-file.")
	}

	// Update the modification time.
	in.attrs.Mtime = time.Now()

	// Ensure that the contents slice is long enough.
	newLen := int(off) + len(p)
	if len(in.contents) < newLen {
		padding := make([]byte, newLen-len(in.contents))
		in.contents = append(in.contents, padding...)
		in.attrs.Size = uint64(newLen)
	}

	// Copy in the data.
	n := copy(in.contents[off:], p)

	// Sanity check.
	if n != len(p) {
		panic(fmt.Sprintf("Unexpected short copy: %v", n))
	}

	return n, nil
}

func (in *inode) Fallocate(mode uint32, offset uint64, length uint64) error {
	if mode != 0 {
		return fuse.ENOSYS
	}
	newSize := int(offset + length)
	if newSize > len(in.contents) {
		padding := make([]byte, newSize-len(in.contents))
		in.contents = append(in.contents, padding...)
		in.attrs.Size = offset + length
	}
	return nil
}

func (in *inode) isDir() bool {
	return in.attrs.Mode.IsDir()
}

func (in *inode) isSymlink() bool {
	return in.attrs.Mode&os.ModeSymlink != 0
}

func (in *inode) isFile() bool {
	return !(in.isDir() || in.isSymlink())
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

// Read from the file's contents. See documentation for ioutil.ReaderAt.
//
// REQUIRES: in.isFile()
func (in *inode) ReadAt(dst []byte, offset int64) (int, error) {
	if !in.isFile() {
		panic("ReadAt called on non-file.")
	}

	// Ensure the offset is in range.
	if offset > int64(len(in.contents)) {
		return 0, io.EOF
	}

	// Read what we can.
	n := copy(dst, in.contents[offset:])
	if n < len(dst) { // INFO: 这里需要判断???
		return n, io.EOF
	}

	return n, nil
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

// Update attributes from non-nil parameters.
func (in *inode) SetAttributes(size *uint64, mode *os.FileMode, mtime *time.Time) {
	// Update the modification time.
	in.attrs.Mtime = time.Now()

	// Truncate?
	if size != nil {
		intSize := int(*size)

		// Update contents.
		if intSize <= len(in.contents) {
			in.contents = in.contents[:intSize]
		} else {
			padding := make([]byte, intSize-len(in.contents))
			in.contents = append(in.contents, padding...)
		}

		// Update attributes.
		in.attrs.Size = *size
	}

	// Change mode?
	if mode != nil {
		in.attrs.Mode = *mode
	}

	// Change mtime?
	if mtime != nil {
		in.attrs.Mtime = *mtime
	}
}
