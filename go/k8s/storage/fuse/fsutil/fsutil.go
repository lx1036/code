package fsutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// Create a temporary file with the same semantics as ioutil.TempFile, but
// ensure that it is unlinked before returning so that it does not persist
// after the process exits.
//
// Warning: this is not production-quality code, and should only be used for
// testing purposes. In particular, there is a race between creating and
// unlinking by name.
func AnonymousFile(dir string) (*os.File, error) {
	// Choose a prefix based on the binary name.
	prefix := path.Base(os.Args[0])

	// Create the file.
	f, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return nil, fmt.Errorf("TempFile: %v", err)
	}

	// Unlink it.
	if err := os.Remove(f.Name()); err != nil {
		return nil, fmt.Errorf("Remove: %v", err)
	}

	return f, nil
}

// Call fdatasync on the supplied file.
//
// REQUIRES: FdatasyncSupported is true.
func Fdatasync(f *os.File) error {
	return fdatasync(f)
}
