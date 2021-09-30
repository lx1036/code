package testing

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"k8s-lx1036/k8s/storage/etcd/storage/backend"
)

var (
	defaultBatchLimit    = 10000
	defaultBatchInterval = 100 * time.Millisecond

	// initialMmapSize is the initial size of the mmapped region. Setting this larger than
	// the potential max db size can prevent writer from blocking reader.
	// This only works for linux.
	initialMmapSize = uint64(10 * 1024 * 1024 * 1024)
)

func NewTmpBackendFromCfg(t testing.TB, bcfg backend.Config) (backend.Backend, string) {
	dir := "tmp"
	os.MkdirAll(dir, 0777)
	tmpPath := filepath.Join(dir, "db.txt")
	bcfg.Path = tmpPath
	return backend.New(bcfg), tmpPath
}

func NewDefaultTmpBackend(t testing.TB) (backend.Backend, string) {
	return NewTmpBackendFromCfg(t, DefaultBackendConfig())
}

func DefaultBackendConfig() backend.Config {
	return backend.Config{
		BatchInterval: defaultBatchInterval,
		BatchLimit:    defaultBatchLimit,
		MmapSize:      initialMmapSize,
	}
}
