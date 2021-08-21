package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
	"k8s-lx1036/k8s/storage/sunfs/cmd/client/fs/fuse/utils"

	"k8s.io/klog/v2"
)

// INFO: Create a file system that stores data and metadata in memory.
//  The supplied UID/GID pair will own the root inode. This file system does no
//  permissions checking, and should therefore be mounted with the default_permissions option.

type MemoryFS struct {
	mutex sync.RWMutex

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

// mkdir -p /tmp/fuse/memoryfs
// go run . --mountpoint=/tmp/fuse/memoryfs
func main() {
	flag.Parse()

	// Mount the file system.
	if *mountPoint == "" {
		klog.Fatalf("You must set --mountpoint.")
	}

	// filesystem server
	server := NewMemoryFS(utils.CurrentUid(), utils.CurrentGid())

	cfg := &fuse.MountConfig{
		ReadOnly: *readOnly,
		FSName:   "memoryfs",
		Subtype:  "memoryfs_subtype",
	}
	if *debug {
		cfg.DebugLogger = log.New(os.Stderr, "fuse: ", 0)
	}

	mountedFileSystem, err := fuse.Mount(*mountPoint, server, cfg)
	if err != nil {
		klog.Fatalf("Mount: %v", err)
	}

	klog.Infof(fmt.Sprintf("fuse mount point %s successfully", *mountPoint))
	// Wait for it to be unmounted.
	if err = mountedFileSystem.Join(context.Background()); err != nil {
		klog.Fatalf("Join: %v", err)
	}
}
