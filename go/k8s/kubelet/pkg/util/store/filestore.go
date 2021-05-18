package store

import (
	"fmt"
	"os"
	"path/filepath"

	utilfs "k8s.io/kubernetes/pkg/util/filesystem"
)

const (
	// Name prefix for the temporary files.
	tmpPrefix = "."
)

// FileStore is an implementation of the Store interface which stores data in files.
type FileStore struct {
	// Absolute path to the base directory for storing data files.
	directoryPath string

	// filesystem to use.
	filesystem utilfs.Filesystem
}

func (f *FileStore) Write(key string, data []byte) error {
	if err := ValidateKey(key); err != nil {
		return err
	}
	if err := ensureDirectory(f.filesystem, f.directoryPath); err != nil {
		return err
	}

	return writeFile(f.filesystem, f.getPathByKey(key), data)
}

// INFO: 一个复杂的函数来写文件
// writeFile writes data to path in a single transaction.
func writeFile(fs utilfs.Filesystem, path string, data []byte) (retErr error) {
	// Create a temporary file in the base directory of `path` with a prefix.
	tmpFile, err := fs.TempFile(filepath.Dir(path), tmpPrefix)
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()
	shouldClose := true

	defer func() {
		// Close the file.
		if shouldClose {
			if err := tmpFile.Close(); err != nil {
				if retErr == nil {
					retErr = fmt.Errorf("close error: %v", err)
				} else {
					retErr = fmt.Errorf("failed to close temp file after error %v; close error: %v", retErr, err)
				}
			}
		}

		// Clean up the temp file on error.
		if retErr != nil && tmpPath != "" {
			if err := removePath(fs, tmpPath); err != nil {
				retErr = fmt.Errorf("failed to remove the temporary file (%q) after error %v; remove error: %v", tmpPath, retErr, err)
			}
		}
	}()

	// Write data.
	if _, err := tmpFile.Write(data); err != nil {
		return err
	}

	// Sync file.
	if err := tmpFile.Sync(); err != nil {
		return err
	}

	// Closing the file before renaming.
	err = tmpFile.Close()
	shouldClose = false
	if err != nil {
		return err
	}

	return fs.Rename(tmpPath, path)
}

func (f *FileStore) Read(key string) ([]byte, error) {
	panic("implement me")
}

func (f *FileStore) Delete(key string) error {
	panic("implement me")
}

func (f *FileStore) List() ([]string, error) {
	panic("implement me")
}

// getPathByKey returns the full path of the file for the key.
func (f *FileStore) getPathByKey(key string) string {
	return filepath.Join(f.directoryPath, key)
}

// NewFileStore returns an instance of *FileStore.
func NewFileStore(path string, fs utilfs.Filesystem) (Store, error) {
	if err := ensureDirectory(fs, path); err != nil {
		return nil, err
	}

	return &FileStore{directoryPath: path, filesystem: fs}, nil
}

// ensureDirectory creates the directory if it does not exist.
func ensureDirectory(fs utilfs.Filesystem, path string) error {
	if _, err := fs.Stat(path); err != nil {
		// MkdirAll returns nil if directory already exists.
		return fs.MkdirAll(path, 0755)
	}
	return nil
}

func removePath(fs utilfs.Filesystem, path string) error {
	if err := fs.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
