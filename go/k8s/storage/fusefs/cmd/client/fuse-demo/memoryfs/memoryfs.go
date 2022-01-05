package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
	"k8s.io/klog/v2"
)

// INFO: Create a file system that stores data and metadata in memory.
//  The supplied UID/GID pair will own the root inode. This file system does no
//  permissions checking, and should therefore be mounted with the default_permissions option.

type MemoryFS struct {
	sync.RWMutex

	fuseutil.NotImplementedFileSystem

	// The UID and GID that every inode receives.
	uid uint32
	gid uint32

	inodes []*inode // GUARDED_BY(mu)

	// A list of inode IDs within inodes available for reuse, not including the
	// reserved IDs less than fuseops.RootInodeID.
	//
	// INVARIANT: This is all and only indices i of 'inodes' such that i >
	// fuseops.RootInodeID and inodes[i] == nil
	freeInodes []fuseops.InodeID // GUARDED_BY(mu)
}

func NewMemoryFS(uid uint32, gid uint32) fuse.Server {
	fs := &MemoryFS{
		inodes: make([]*inode, fuseops.RootInodeID+1),
		uid:    uid,
		gid:    gid,
	}

	fs.inodes[fuseops.RootInodeID] = newInode(fuseops.InodeAttributes{
		Mode: 0700 | os.ModeDir,
		Uid:  uid,
		Gid:  gid,
	})

	return fuseutil.NewFileSystemServer(fs)
}

func (fs *MemoryFS) getInodeOrDie(id fuseops.InodeID) *inode {
	inode := fs.inodes[id]
	if inode == nil {
		panic(fmt.Sprintf("Unknown inode: %v", id))
	}

	//klog.Infof(fmt.Sprintf("[getInodeOrDie]inodeID:%d for %+v", id, *inode))
	return inode
}

func (fs *MemoryFS) StatFS(ctx context.Context, op *fuseops.StatFSOp) error {
	return nil
}

func (fs *MemoryFS) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	klog.Infof(fmt.Sprintf("[LookUpInode]%+v", *op))

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

	return nil
}

func (fs *MemoryFS) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	klog.Infof(fmt.Sprintf("[OpenDir]%+v", *op))

	currentInode := fs.getInodeOrDie(op.Inode)
	if !currentInode.isDir() {
		panic("Found non-dir.")
	}

	return nil
}

func (fs *MemoryFS) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	// Grab the directory.
	currentInode := fs.getInodeOrDie(op.Inode)
	// Serve the request.
	op.BytesRead = currentInode.ReadDir(op.Dst, int(op.Offset))

	klog.Infof(fmt.Sprintf("[ReadDir]inodeID:%d, handleID:%d, Offset:%d, BytesRead:%d",
		op.Inode, op.Handle, op.Offset, op.BytesRead))
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

	klog.Infof(fmt.Sprintf("[ReadFile]inodeID:%d, handleID:%d, Offset:%d, BytesRead:%d",
		op.Inode, op.Handle, op.Offset, op.BytesRead))
	// Don't return EOF errors; we just indicate EOF to fuse using a short read.
	if err == io.EOF {
		return nil
	}

	return err
}

// Allocate a new inode, assigning it an ID that is not in use.
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

func (fs *MemoryFS) SetInodeAttributes(ctx context.Context, op *fuseops.SetInodeAttributesOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	klog.Infof(fmt.Sprintf("[SetInodeAttributes]%+v", *op))

	var err error
	if op.Size != nil && op.Handle == nil && *op.Size != 0 {
		// require that truncate to non-zero has to be ftruncate()
		// but allow open(O_TRUNC)
		err = syscall.EBADF
	}

	// Grab the inode.
	inode := fs.getInodeOrDie(op.Inode)
	// Handle the request.
	inode.SetAttributes(op.Size, op.Mode, op.Mtime)

	// Fill in the response.
	op.Attributes = inode.attrs

	// We don't spontaneously mutate, so the kernel can cache as long as it wants
	// (since it also handles invalidation).
	op.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)

	return err
}

func (fs *MemoryFS) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	klog.Infof(fmt.Sprintf("[GetInodeAttributes]%+v", *op))

	// 先查找当前文件
	currentInode := fs.getInodeOrDie(op.Inode)
	// Fill in the response.
	op.Attributes = currentInode.attrs
	op.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)

	return nil
}

// MkDir `mkdir abc`
// LookUpInode -> MkDir
func (fs *MemoryFS) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	klog.Infof(fmt.Sprintf("[MkDir]%+v", *op))

	// Grab the parent, which we will update shortly.
	parent := fs.getInodeOrDie(op.Parent)
	// Ensure that the name doesn't already exist, so we don't wind up with a duplicate.
	_, _, exists := parent.LookUpChild(op.Name)
	if exists {
		return fuse.EEXIST
	}

	childID, child := fs.allocateInode(fuseops.InodeAttributes{
		Nlink: 1,
		Mode:  op.Mode,
		Uid:   fs.uid,
		Gid:   fs.gid,
	})
	// Add an entry in the parent.
	parent.AddChild(childID, op.Name, fuseutil.DT_Directory)
	op.Entry.Child = childID
	op.Entry.Attributes = child.attrs
	op.Entry.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)
	op.Entry.EntryExpiration = op.Entry.AttributesExpiration

	return nil
}

func (fs *MemoryFS) MkNode(ctx context.Context, op *fuseops.MkNodeOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	klog.Infof(fmt.Sprintf("[MkNode]%+v", *op))

	var err error
	op.Entry, err = fs.createFile(op.Parent, op.Name, op.Mode)
	return err
}

// LOCKS_REQUIRED(fs.)
func (fs *MemoryFS) createFile(parentID fuseops.InodeID, name string, mode os.FileMode) (fuseops.ChildInodeEntry, error) {
	// Grab the parent, which we will update shortly.
	parent := fs.getInodeOrDie(parentID)

	// Ensure that the name doesn't already exist, so we don't wind up with a
	// duplicate.
	_, _, exists := parent.LookUpChild(name)
	if exists {
		return fuseops.ChildInodeEntry{}, fuse.EEXIST
	}

	// Set up attributes for the child.
	now := time.Now()
	childAttrs := fuseops.InodeAttributes{
		Nlink:  1,
		Mode:   mode,
		Atime:  now,
		Mtime:  now,
		Ctime:  now,
		Crtime: now,
		Uid:    fs.uid,
		Gid:    fs.gid,
	}

	// Allocate a child.
	childID, child := fs.allocateInode(childAttrs)

	// Add an entry in the parent.
	parent.AddChild(childID, name, fuseutil.DT_File)

	// Fill in the response entry.
	var entry fuseops.ChildInodeEntry
	entry.Child = childID
	entry.Attributes = child.attrs

	// We don't spontaneously mutate, so the kernel can cache as long as it wants
	// (since it also handles invalidation).
	entry.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)
	entry.EntryExpiration = entry.AttributesExpiration

	return entry, nil
}

func (fs *MemoryFS) CreateFile(ctx context.Context, op *fuseops.CreateFileOp) (err error) {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	op.Entry, err = fs.createFile(op.Parent, op.Name, op.Mode)
	return err
}

func (fs *MemoryFS) CreateSymlink(ctx context.Context, op *fuseops.CreateSymlinkOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	// Grab the parent, which we will update shortly.
	parent := fs.getInodeOrDie(op.Parent)

	// Ensure that the name doesn't already exist, so we don't wind up with a
	// duplicate.
	_, _, exists := parent.LookUpChild(op.Name)
	if exists {
		return fuse.EEXIST
	}

	// Set up attributes from the child.
	now := time.Now()
	childAttrs := fuseops.InodeAttributes{
		Nlink:  1,
		Mode:   0444 | os.ModeSymlink,
		Atime:  now,
		Mtime:  now,
		Ctime:  now,
		Crtime: now,
		Uid:    fs.uid,
		Gid:    fs.gid,
	}

	// Allocate a child.
	childID, child := fs.allocateInode(childAttrs)

	// Set up its target.
	child.target = op.Target

	// Add an entry in the parent.
	parent.AddChild(childID, op.Name, fuseutil.DT_Link)

	// Fill in the response entry.
	op.Entry.Child = childID
	op.Entry.Attributes = child.attrs

	// We don't spontaneously mutate, so the kernel can cache as long as it wants
	// (since it also handles invalidation).
	op.Entry.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)
	op.Entry.EntryExpiration = op.Entry.AttributesExpiration

	return nil
}

func (fs *MemoryFS) CreateLink(ctx context.Context, op *fuseops.CreateLinkOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	// Grab the parent, which we will update shortly.
	parent := fs.getInodeOrDie(op.Parent)

	// Ensure that the name doesn't already exist, so we don't wind up with a
	// duplicate.
	_, _, exists := parent.LookUpChild(op.Name)
	if exists {
		return fuse.EEXIST
	}

	// Get the target inode to be linked
	target := fs.getInodeOrDie(op.Target)

	// Update the attributes
	now := time.Now()
	target.attrs.Nlink++
	target.attrs.Ctime = now

	// Add an entry in the parent.
	parent.AddChild(op.Target, op.Name, fuseutil.DT_File)

	// Return the response.
	op.Entry.Child = op.Target
	op.Entry.Attributes = target.attrs

	// We don't spontaneously mutate, so the kernel can cache as long as it wants
	// (since it also handles invalidation).
	op.Entry.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)
	op.Entry.EntryExpiration = op.Entry.AttributesExpiration

	return nil
}

func (fs *MemoryFS) Rename(ctx context.Context, op *fuseops.RenameOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	// Ask the old parent for the child's inode ID and type.
	oldParent := fs.getInodeOrDie(op.OldParent)
	childID, childType, ok := oldParent.LookUpChild(op.OldName)

	if !ok {
		return fuse.ENOENT
	}

	// If the new name exists already in the new parent, make sure it's not a
	// non-empty directory, then delete it.
	newParent := fs.getInodeOrDie(op.NewParent)
	existingID, _, ok := newParent.LookUpChild(op.NewName)
	if ok {
		existing := fs.getInodeOrDie(existingID)

		var buf [4096]byte
		if existing.isDir() && existing.ReadDir(buf[:], 0) > 0 {
			return fuse.ENOTEMPTY
		}

		newParent.RemoveChild(op.NewName)
	}

	// Link the new name.
	newParent.AddChild(
		childID,
		op.NewName,
		childType)

	// Finally, remove the old name from the old parent.
	oldParent.RemoveChild(op.OldName)

	return nil
}

func (fs *MemoryFS) RmDir(ctx context.Context, op *fuseops.RmDirOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	// Grab the parent, which we will update shortly.
	parent := fs.getInodeOrDie(op.Parent)

	// Find the child within the parent.
	childID, _, ok := parent.LookUpChild(op.Name)
	if !ok {
		return fuse.ENOENT
	}

	// Grab the child.
	child := fs.getInodeOrDie(childID)

	// Make sure the child is empty.
	if child.Len() != 0 {
		return fuse.ENOTEMPTY
	}

	// Remove the entry within the parent.
	parent.RemoveChild(op.Name)

	// Mark the child as unlinked.
	child.attrs.Nlink--

	return nil
}

func (fs *MemoryFS) Unlink(ctx context.Context, op *fuseops.UnlinkOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	// Grab the parent, which we will update shortly.
	parent := fs.getInodeOrDie(op.Parent)

	// Find the child within the parent.
	childID, _, ok := parent.LookUpChild(op.Name)
	if !ok {
		return fuse.ENOENT
	}

	// Grab the child.
	child := fs.getInodeOrDie(childID)

	// Remove the entry within the parent.
	parent.RemoveChild(op.Name)

	// Mark the child as unlinked.
	child.attrs.Nlink--

	return nil
}

func (fs *MemoryFS) OpenFile(ctx context.Context, op *fuseops.OpenFileOp) error {
	if op.OpContext.Pid == 0 {
		// OpenFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	// We don't mutate spontaneosuly, so if the VFS layer has asked for an
	// inode that doesn't exist, something screwed up earlier (a lookup, a
	// cache invalidation, etc.).
	inode := fs.getInodeOrDie(op.Inode)

	if !inode.isFile() {
		panic("Found non-file.")
	}

	return nil
}

func (fs *MemoryFS) WriteFile(ctx context.Context, op *fuseops.WriteFileOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	// Find the inode in question.
	inode := fs.getInodeOrDie(op.Inode)

	// Serve the request.
	_, err := inode.WriteAt(op.Data, op.Offset)

	return err
}

func (fs *MemoryFS) FlushFile(ctx context.Context, op *fuseops.FlushFileOp) (err error) {
	if op.OpContext.Pid == 0 {
		// FlushFileOp should have a valid pid in context.
		return fuse.EINVAL
	}
	return
}

func (fs *MemoryFS) ReadSymlink(ctx context.Context, op *fuseops.ReadSymlinkOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	// Find the inode in question.
	inode := fs.getInodeOrDie(op.Inode)

	// Serve the request.
	op.Target = inode.target

	return nil
}

func (fs *MemoryFS) GetXattr(ctx context.Context, op *fuseops.GetXattrOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	inode := fs.getInodeOrDie(op.Inode)
	if value, ok := inode.xattrs[op.Name]; ok {
		op.BytesRead = len(value)
		if len(op.Dst) >= len(value) {
			copy(op.Dst, value)
		} else if len(op.Dst) != 0 {
			return syscall.ERANGE
		}
	} else {
		return fuse.ENOATTR
	}

	return nil
}

func (fs *MemoryFS) ListXattr(ctx context.Context, op *fuseops.ListXattrOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	inode := fs.getInodeOrDie(op.Inode)

	dst := op.Dst[:]
	for key := range inode.xattrs {
		keyLen := len(key) + 1

		if len(dst) >= keyLen {
			copy(dst, key)
			dst = dst[keyLen:]
		} else if len(op.Dst) != 0 {
			return syscall.ERANGE
		}
		op.BytesRead += keyLen
	}

	return nil
}

func (fs *MemoryFS) RemoveXattr(ctx context.Context, op *fuseops.RemoveXattrOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()
	inode := fs.getInodeOrDie(op.Inode)

	if _, ok := inode.xattrs[op.Name]; ok {
		delete(inode.xattrs, op.Name)
	} else {
		return fuse.ENOATTR
	}
	return nil
}

func (fs *MemoryFS) SetXattr(ctx context.Context, op *fuseops.SetXattrOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()
	inode := fs.getInodeOrDie(op.Inode)

	_, ok := inode.xattrs[op.Name]

	switch op.Flags {
	case unix.XATTR_CREATE:
		if ok {
			return fuse.EEXIST
		}
	case unix.XATTR_REPLACE:
		if !ok {
			return fuse.ENOATTR
		}
	}

	value := make([]byte, len(op.Value))
	copy(value, op.Value)
	inode.xattrs[op.Name] = value
	return nil
}

func (fs *MemoryFS) Fallocate(ctx context.Context, op *fuseops.FallocateOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()
	inode := fs.getInodeOrDie(op.Inode)
	inode.Fallocate(op.Mode, op.Offset, op.Length)
	return nil
}

var (
	mountPoint = flag.String("mountpoint", "", "Path to mount point.")
	readOnly   = flag.Bool("read_only", false, "Mount in read-only mode.")
	debug      = flag.Bool("debug", false, "Enable debug logging.")
)

// INFO: 可以直接在 mac 上运行
//  mkdir -p /tmp/fuse/memoryfs
//  go run . --mountpoint=/tmp/fuse/memoryfs
//  `umount globalmount` in mac
//  `fusermount -u globalmount` in linux
func main() {
	flag.Parse()

	if len(*mountPoint) == 0 {
		klog.Fatalf("You must set --mountpoint.")
	}

	// filesystem server
	value, _ := user.Current()
	uid, _ := strconv.ParseUint(value.Uid, 10, 32)
	gid, _ := strconv.ParseUint(value.Gid, 10, 32)
	server := NewMemoryFS(uint32(uid), uint32(gid))

	cfg := &fuse.MountConfig{
		ReadOnly:   *readOnly,
		FSName:     "memoryfs",
		Subtype:    "fuse",
		VolumeName: "memoryfs", // OS X only
	}
	if *debug {
		cfg.DebugLogger = log.New(os.Stderr, "fuse: ", 0)
	}

	mountPath, _ := filepath.Abs(*mountPoint)
	mountedFileSystem, err := fuse.Mount(mountPath, server, cfg)
	if err != nil {
		klog.Fatalf("Mount: %v", err)
	}

	klog.Infof(fmt.Sprintf("fuse mount point %s successfully", mountPath))
	if err = mountedFileSystem.Join(context.Background()); err != nil {
		klog.Fatalf("Join: %v", err)
	}
}
