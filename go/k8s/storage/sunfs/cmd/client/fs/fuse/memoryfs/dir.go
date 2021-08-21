package main

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"syscall"
	"time"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
)

func (fs *MemoryFS) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) error {
	return nil
}

func (fs *MemoryFS) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// 先查找当前文件
	currentInode := fs.getInodeOrDie(op.Inode)
	// Fill in the response.
	op.Attributes = currentInode.attrs
	// We don't spontaneously mutate, so the kernel can cache as long as it wants
	// (since it also handles invalidation).
	op.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)

	klog.Infof(fmt.Sprintf("[GetInodeAttributes]inode id %d, uid: %d, gid: %d", op.Inode, op.Attributes.Uid, op.Attributes.Gid))
	return nil
}

func (fs *MemoryFS) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// Grab the parent directory.
	parentInode := fs.getInodeOrDie(op.Parent)
	// Does the directory have an entry with the given name?
	childID, _, ok := parentInode.LookUpChild(op.Name)
	if !ok {
		return fuse.ENOENT
	}

	klog.Infof(fmt.Sprintf("[LookUpInode]childID: %d", childID))

	// Grab the child.
	child := fs.getInodeOrDie(childID)

	// Fill in the response.
	op.Entry.Child = childID
	op.Entry.Attributes = child.attrs
	// We don't spontaneously mutate, so the kernel can cache as long as it wants
	// (since it also handles invalidation).
	op.Entry.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)
	op.Entry.EntryExpiration = op.Entry.AttributesExpiration

	klog.Infof(fmt.Sprintf("[LookUpInode]parent id %d, Name %s", op.Parent, op.Name))
	return nil
}

func (fs *MemoryFS) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// We don't mutate spontaneosuly, so if the VFS layer has asked for an
	// inode that doesn't exist, something screwed up earlier (a lookup, a
	// cache invalidation, etc.).
	currentInode := fs.getInodeOrDie(op.Inode)
	if !currentInode.isDir() {
		panic("Found non-dir.")
	}

	klog.Infof(fmt.Sprintf("[OpenDir]inode id: %d", op.Inode))
	return nil
}

func (fs *MemoryFS) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// Grab the directory.
	currentInode := fs.getInodeOrDie(op.Inode)
	// Serve the request.
	op.BytesRead = currentInode.ReadDir(op.Dst, int(op.Offset))

	klog.Infof(fmt.Sprintf("[ReadDir]inodeID:%d, Offset: %d, BytesRead: %d",
		op.Inode, op.Offset, op.BytesRead))
	return nil
}

func (fs *MemoryFS) GetXattr(ctx context.Context, op *fuseops.GetXattrOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// Grab the directory.
	currentInode := fs.getInodeOrDie(op.Inode)
	if value, ok := currentInode.xattrs[op.Name]; ok {
		// op.Dst 长度必须大于 value 长度, 见 GetXattrOp.Dst 字段注解
		op.BytesRead = len(value)
		if len(op.Dst) >= len(value) {
			copy(op.Dst, value)
		} else if len(op.Dst) != 0 {
			return syscall.ERANGE
		}
	} else {
		return fuse.ENOATTR
	}

	klog.Infof(fmt.Sprintf("[GetXattr]inodeID:%d, Name:%s, Dst: %s", op.Inode, op.Name, string(op.Dst)))
	return nil
}

func (fs *MemoryFS) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	//INFO: 检查 parent 里是否已经存在该 op.Name
	parentInode := fs.getInodeOrDie(op.Parent)
	_, _, exists := parentInode.LookUpChild(op.Name)
	if exists {
		return fuse.EEXIST
	}

	// Set up attributes from the child.
	childAttrs := fuseops.InodeAttributes{
		Nlink: 1,
		Mode:  op.Mode,
		Uid:   fs.uid,
		Gid:   fs.gid,
	}
	// Allocate a child.
	childID, child := fs.allocateInode(childAttrs)
	klog.Infof(fmt.Sprintf("[MkDir]allocateInode childID: %d", childID))
	// Add an entry in the parent.
	parentInode.AddChild(childID, op.Name, fuseutil.DT_Directory)
	// Fill in the response.
	op.Entry.Child = childID
	op.Entry.Attributes = child.attrs
	// We don't spontaneously mutate, so the kernel can cache as long as it wants
	// (since it also handles invalidation).
	op.Entry.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)
	op.Entry.EntryExpiration = op.Entry.AttributesExpiration

	klog.Infof(fmt.Sprintf("[MkDir]Parent:%d, Name:%s, Mode:%s", op.Parent, op.Name, op.Mode.String()))
	return nil
}
