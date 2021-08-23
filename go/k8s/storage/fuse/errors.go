package fuse

import "syscall"

const (
	// Errors corresponding to kernel error numbers. These may be treated
	// specially by Connection.Reply.
	EEXIST    = syscall.EEXIST
	EINVAL    = syscall.EINVAL
	EIO       = syscall.EIO
	ENOATTR   = syscall.ENODATA
	ENOENT    = syscall.ENOENT
	ENOSYS    = syscall.ENOSYS
	ENOTDIR   = syscall.ENOTDIR
	ENOTEMPTY = syscall.ENOTEMPTY
)
