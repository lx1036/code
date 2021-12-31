package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"

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
		Subtype:  "memoryfs_subtype",
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
