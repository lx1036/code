package fsutil

import (
	"os"
	"syscall"
)

const FdatasyncSupported = true

func fdatasync(f *os.File) error {
	return syscall.Fdatasync(int(f.Fd()))
}
