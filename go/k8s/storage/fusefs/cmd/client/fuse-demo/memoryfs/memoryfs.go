package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
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

	// The collection of live inodes, indexed by ID. IDs of free inodes that may
	// be re-used have nil entries. No ID less than fuseops.RootInodeID is ever
	// used.
	//
	// All inodes are protected by the file system mutex.
	//
	// INVARIANT: For each inode in, in.CheckInvariants() does not panic.
	// INVARIANT: len(inodes) > fuseops.RootInodeID
	// INVARIANT: For all i < fuseops.RootInodeID, inodes[i] == nil
	// INVARIANT: inodes[fuseops.RootInodeID] != nil
	// INVARIANT: inodes[fuseops.RootInodeID].isDir()
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

	// INFO: /tmp/fuse/memoryfs/ 目录是根目录
	//  Set up the root inode.
	rootAttrs := fuseops.InodeAttributes{
		Mode: 0700 | os.ModeDir,
		Uid:  uid,
		Gid:  gid,
	}
	fs.inodes[fuseops.RootInodeID] = newInode(rootAttrs)

	return fuseutil.NewFileSystemServer(fs)
}

func (fs *MemoryFS) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) error {
	return nil
}

func (fs *MemoryFS) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	klog.Infof(fmt.Sprintf("[GetInodeAttributes]inodeID:%d", op.Inode))

	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

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

	fs.Lock()
	defer fs.Unlock()

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

	fs.Lock()
	defer fs.Unlock()

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

	fs.Lock()
	defer fs.Unlock()

	// Grab the directory.
	currentInode := fs.getInodeOrDie(op.Inode)
	// Serve the request.
	op.BytesRead = currentInode.ReadDir(op.Dst, int(op.Offset))

	klog.Infof(fmt.Sprintf("[ReadDir]inodeID:%d, Offset: %d, BytesRead: %d",
		op.Inode, op.Offset, op.BytesRead))
	return nil
}

/*func (fs *MemoryFS) GetXattr(ctx context.Context, op *fuseops.GetXattrOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

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

	fs.Lock()
	defer fs.Unlock()

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

	klog.Infof(fmt.Sprintf("[CreateFile]create filename %s, parent inodeID %d", op.Name, op.Parent))
	return nil
}
*/
func (fs *MemoryFS) ReadFile(ctx context.Context, op *fuseops.ReadFileOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	fs.Lock()
	defer fs.Unlock()

	klog.Infof(fmt.Sprintf("[ReadFile]dst length %d", len(op.Dst)))

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

	// Mount the file system.
	if *mountPoint == "" {
		klog.Fatalf("You must set --mountpoint.")
	}

	// filesystem server
	value, _ := user.Current()
	uid, _ := strconv.ParseUint(value.Uid, 10, 32)
	gid, _ := strconv.ParseUint(value.Gid, 10, 32)
	server := NewMemoryFS(uint32(uid), uint32(gid))

	cfg := &fuse.MountConfig{
		ReadOnly: *readOnly,
		FSName:   "memoryfs",
		Subtype:  "fuse",
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
