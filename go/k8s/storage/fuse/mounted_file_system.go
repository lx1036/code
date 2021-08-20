package fuse

import "context"

// MountedFileSystem represents the status of a mount operation, with a method
// that waits for unmounting.
type MountedFileSystem struct {
	dir string

	// The result to return from Join. Not valid until the channel is closed.
	joinStatus          error
	joinStatusAvailable chan struct{}
}

// Dir returns the directory on which the file system is mounted (or where we
// attempted to mount it.)
func (mfs *MountedFileSystem) Dir() string {
	return mfs.dir
}

// Join blocks until a mounted file system has been unmounted. It does not
// return successfully until all ops read from the connection have been
// responded to (i.e. the file system server has finished processing all
// in-flight ops).
//
// The return value will be non-nil if anything unexpected happened while
// serving. May be called multiple times.
func (mfs *MountedFileSystem) Join(ctx context.Context) error {
	select {
	case <-mfs.joinStatusAvailable:
		return mfs.joinStatus
	case <-ctx.Done():
		return ctx.Err()
	}
}
