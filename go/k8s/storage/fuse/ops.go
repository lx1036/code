package fuse

import (
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fuse/internal/fusekernel"
)

// A sentinel used for unknown ops. The user is expected to respond with a
// non-nil error.
type unknownOp struct {
	OpCode uint32
	Inode  fuseops.InodeID
}

// Causes us to cancel the associated context.
type interruptOp struct {
	FuseID uint64
}

// Required in order to mount on Linux and OS X.
type initOp struct {
	// In
	Kernel fusekernel.Protocol

	// In/out
	Flags fusekernel.InitFlags

	// Out
	Library       fusekernel.Protocol
	MaxReadahead  uint32
	MaxBackground uint16
	MaxWrite      uint32
	MaxPages      uint16
}
