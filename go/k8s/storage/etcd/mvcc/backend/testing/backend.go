package testing

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.etcd.io/etcd/server/v3/mvcc/backend"
	"go.uber.org/zap/zaptest"
)

var (
	defaultBatchLimit    = 10000
	defaultBatchInterval = 100 * time.Millisecond

	// initialMmapSize is the initial size of the mmapped region. Setting this larger than
	// the potential max db size can prevent writer from blocking reader.
	// This only works for linux.
	initialMmapSize = uint64(10 * 1024 * 1024 * 1024)
)

func NewTmpBackendFromCfg(t testing.TB, bcfg backend.BackendConfig) (backend.Backend, string) {
	dir := "tmp"
	os.MkdirAll(dir, 0777)
	tmpPath := filepath.Join(dir, "db.txt")
	bcfg.Path = tmpPath
	bcfg.Logger = zaptest.NewLogger(t)
	return backend.New(bcfg), tmpPath
}

func NewDefaultTmpBackend(t testing.TB) (backend.Backend, string) {
	return NewTmpBackendFromCfg(t, DefaultBackendConfig())
}

func DefaultBackendConfig() backend.BackendConfig {
	return backend.BackendConfig{
		BatchInterval: defaultBatchInterval,
		BatchLimit:    defaultBatchLimit,
		MmapSize:      initialMmapSize,
	}
}
