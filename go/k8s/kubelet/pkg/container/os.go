package container

import (
	"os"
	"time"
)

// OSInterface collects system level operations that need to be mocked out
// during tests.
type OSInterface interface {
	MkdirAll(path string, perm os.FileMode) error
	Symlink(oldname string, newname string) error
	Stat(path string) (os.FileInfo, error)
	Remove(path string) error
	RemoveAll(path string) error
	Create(path string) (*os.File, error)
	Chmod(path string, perm os.FileMode) error
	Hostname() (name string, err error)
	Chtimes(path string, atime time.Time, mtime time.Time) error
	Pipe() (r *os.File, w *os.File, err error)
	ReadDir(dirname string) ([]os.FileInfo, error)
	Glob(pattern string) ([]string, error)
	Open(name string) (*os.File, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Rename(oldpath, newpath string) error
}
