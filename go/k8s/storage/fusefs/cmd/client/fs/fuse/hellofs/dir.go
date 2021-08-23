package main

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"

	"k8s.io/klog/v2"
)

// MkDir Create a directory inode as a child of an existing directory inode.
// The kernel sends this in response to a mkdir(2) call.
func (fs *helloFS) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {
	klog.Infof(fmt.Sprintf("[MkDir]mkdir %+v", *op))

	return nil
}

func (fs *helloFS) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	// Allow opening any directory.
	return nil
}

// INFO: `ll /tmp/fuse/hellofs` read dir
func (fs *helloFS) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	// Find the info for this inode.
	info, ok := gInodeInfo[op.Inode]
	if !ok {
		return fuse.ENOENT
	}

	klog.Infof(fmt.Sprintf("ReadDirOp Inode: %+v, InodeInfo: %+v", op.Inode, info))

	if !info.dir {
		return fuse.EIO
	}

	entries := info.children

	// Grab the range of interest.
	if op.Offset > fuseops.DirOffset(len(entries)) {
		return fuse.EIO
	}

	entries = entries[op.Offset:]

	// Resume at the specified offset into the array.
	for _, e := range entries {
		n := fuseutil.WriteDirent(op.Dst[op.BytesRead:], e)
		if n == 0 {
			break
		}

		op.BytesRead += n
	}

	return nil
}

func (fs *helloFS) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	// Find the info for the parent.
	parentInfo, ok := gInodeInfo[op.Parent]
	if !ok {
		return fuse.ENOENT
	}

	// Find the child within the parent.
	childInode, err := findChildInode(op.Name, parentInfo.children)
	if err != nil {
		return err
	}

	// Copy over information.
	op.Entry.Child = childInode
	op.Entry.Attributes = gInodeInfo[childInode].attributes

	// Patch attributes.
	fs.patchAttributes(&op.Entry.Attributes)

	klog.Infof(fmt.Sprintf("[LookUpInode]"))
	return nil
}

func (fs *helloFS) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	// Find the info for this inode.
	info, ok := gInodeInfo[op.Inode]
	if !ok {
		return fuse.ENOENT
	}

	// Copy over its attributes.
	op.Attributes = info.attributes

	// Patch attributes.
	fs.patchAttributes(&op.Attributes)

	return nil
}

func (fs *helloFS) patchAttributes(attr *fuseops.InodeAttributes) {
	now := time.Now()
	attr.Atime = now
	attr.Mtime = now
	attr.Crtime = now
}
