package fs

import (
	"sync"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s-lx1036/k8s/storage/fuse/fuseops"
)

type HandleCache struct {
	sync.RWMutex
	handles      map[fuseops.HandleID]*DirHandle
	currHandleID uint64
}

type DirHandle struct {
	// Imutable
	ino uint64

	// For directory handles only
	lock    sync.Mutex
	entries []proto.Dentry
}
