package lock

import (
	"errors"
	"fmt"
	"k8s-lx1036/k8s/bpf/toa/internal/sysconn"
	"os"

	"golang.org/x/sys/unix"
)

type File struct {
	*os.File
	how int
}

func Exclusive(file *os.File) *File {
	return &File{file, unix.LOCK_EX}
}

// OpenLockedExclusive opens the given path and acquires an exclusive lock.
func OpenLockedExclusive(path string) (*File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	lock := Exclusive(file)
	lock.Lock()
	return lock, nil
}

// Lock implements sync.Locker.
//
// It panics if the underlying syscalls return an error.
func (fl *File) Lock() {
	if err := fl.flock(fl.how); err != nil {
		panic(err.Error())
	}
}

func (fl *File) flock(how int) error {
	err := sysconn.Control(fl.File, func(fd int) (err error) {
		for {
			err = unix.Flock(int(fd), how)
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return
		}
	})

	if err != nil {
		return fmt.Errorf("flock: %w", err)
	}
	return nil
}
