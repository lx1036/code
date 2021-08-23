package main

import (
	"context"
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"time"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
)

func (fs *MemoryFS) CreateFile(ctx context.Context, op *fuseops.CreateFileOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	// 先查找parent文件
	parent := fs.getInodeOrDie(op.Parent)
	_, _, exists := parent.LookUpChild(op.Name) // op.Name 是要创建文件的 name
	if exists {
		return fuse.EEXIST
	}

	// Set up attributes for the child.
	now := time.Now()
	childAttrs := fuseops.InodeAttributes{
		Nlink:  1,
		Mode:   op.Mode,
		Atime:  now,
		Mtime:  now,
		Ctime:  now,
		Crtime: now,
		Uid:    fs.uid,
		Gid:    fs.gid,
	}
	// Allocate a child.
	childID, child := fs.allocateInode(childAttrs)
	klog.Infof(fmt.Sprintf("[CreateFile]allocateInode childID: %d", childID))
	// Add an entry in the parent.
	parent.AddChild(childID, op.Name, fuseutil.DT_File)

	// Fill in the response entry.
	op.Entry.Child = childID
	op.Entry.Attributes = child.attrs
	// We don't spontaneously mutate, so the kernel can cache as long as it wants
	// (since it also handles invalidation).
	op.Entry.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)
	op.Entry.EntryExpiration = op.Entry.AttributesExpiration

	klog.Infof(fmt.Sprintf("[CreateFile]create filename %s, parent %d", op.Name, op.Parent))
	return nil
}

func (fs *MemoryFS) ReadFile(ctx context.Context, op *fuseops.ReadFileOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	var err error
	currentInode := fs.getInodeOrDie(op.Inode)
	op.BytesRead, err = currentInode.ReadAt(op.Dst, op.Offset)

	// Don't return EOF errors; we just indicate EOF to fuse using a short read.
	if err == io.EOF {
		return nil
	}

	return err
}

// Find the given inode. Panic if it doesn't exist.
//
// LOCKS_REQUIRED(fs.mu)
func (fs *MemoryFS) getInodeOrDie(id fuseops.InodeID) *inode {
	node := fs.inodes[id]
	if node == nil {
		panic(fmt.Sprintf("Unknown inode: %v", id))
	}

	return node
}

// Allocate a new inode, assigning it an ID that is not in use.
//
// LOCKS_REQUIRED(fs.mu)
func (fs *MemoryFS) allocateInode(attrs fuseops.InodeAttributes) (id fuseops.InodeID, inode *inode) {
	// Create the inode.
	inode = newInode(attrs)

	// Re-use a free ID if possible. Otherwise mint a new one.
	numFree := len(fs.freeInodes)
	if numFree != 0 {
		id = fs.freeInodes[numFree-1]
		fs.freeInodes = fs.freeInodes[:numFree-1]
		fs.inodes[id] = inode
	} else {
		id = fuseops.InodeID(len(fs.inodes))
		fs.inodes = append(fs.inodes, inode)
	}

	return id, inode
}
