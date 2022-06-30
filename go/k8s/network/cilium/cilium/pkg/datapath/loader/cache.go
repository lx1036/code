package loader

import (
	"os"
	"path/filepath"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/defaults"
)

// RestoreTemplates populates the object cache from templates on the filesystem
// at the specified path.
// Delete /var/run/cilium/state/templates
func RestoreTemplates(stateDir string) error {
	// Simplest implementation: Just garbage-collect everything.
	// In future we should make this smarter.
	path := filepath.Join(stateDir, defaults.TemplatesDir)
	err := os.RemoveAll(path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return &os.PathError{
		Op:   "failed to remove old BPF templates",
		Path: path,
		Err:  err,
	}
}
