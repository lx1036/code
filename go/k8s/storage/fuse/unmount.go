package fuse

// Unmount attempts to unmount the file system whose mount point is the
// supplied directory.
func Unmount(dir string) error {
	return unmount(dir)
}
