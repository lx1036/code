package fsutil

import "os"

const FdatasyncSupported = false

func fdatasync(f *os.File) error {
	panic("We require FdatasyncSupported be true.")
}
