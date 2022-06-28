package bpf

import (
	"path/filepath"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/defaults"
)

var (
	// Path to where bpffs is mounted
	mapRoot = defaults.DefaultMapRoot

	// Prefix for all maps (default: tc/globals)
	mapPrefix = defaults.DefaultMapPrefix
)

// MapPath returns a path for a BPF map with a given name.
func MapPath(name string) string { // /sys/fs/bpf/tc/globals
	return filepath.Join(mapRoot, mapPrefix, name)
}
