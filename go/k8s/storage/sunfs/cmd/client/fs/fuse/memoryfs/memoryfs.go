package memoryfs

import (
	"sync"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
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
}

func NewMemoryFS(uid uint32, gid uint32) fuse.Server {
	fs := &MemoryFS{
		inodes: make([]*inode, fuseops.RootInodeID+1),
		uid:    uid,
		gid:    gid,
	}
}
