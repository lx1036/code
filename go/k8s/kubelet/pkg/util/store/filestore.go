package store

import utilfs "k8s.io/kubernetes/pkg/util/filesystem"

// FileStore is an implementation of the Store interface which stores data in files.
type FileStore struct {
	// Absolute path to the base directory for storing data files.
	directoryPath string

	// filesystem to use.
	filesystem utilfs.Filesystem
}

func (f *FileStore) Write(key string, data []byte) error {
	panic("implement me")
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
