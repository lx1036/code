package bpf

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/defaults"
)

var (
	once sync.Once
	// Set to true on first get request to detect misorder
	lockedDown = false

	// Path to where bpffs is mounted
	mapRoot = defaults.DefaultMapRoot

	// Prefix for all maps (default: tc/globals)
	mapPrefix = defaults.DefaultMapPrefix
)

func lockDown() {
	lockedDown = true
}

func GetMapRoot() string { // "/sys/fs/bpf"
	once.Do(lockDown)
	return mapRoot
}

// MapPath returns a path for a BPF map with a given name.
func MapPath(name string) string { // /sys/fs/bpf/tc/globals
	return filepath.Join(mapRoot, mapPrefix, name)
}

// LocalMapPath returns the path for a BPF map that is local to the specified ID.
func LocalMapPath(name string, id uint16) string {
	return MapPath(LocalMapName(name, id))
}

// LocalMapName returns the name for a BPF map that is local to the specified ID.
func LocalMapName(name string, id uint16) string {
	return fmt.Sprintf("%s%05d", name, id)
}

// Environment returns a list of environment variables which are needed to make
// BPF programs and tc aware of the actual BPFFS mount path.
func Environment() []string {
	return append(
		os.Environ(),
		fmt.Sprintf("CILIUM_BPF_MNT=%s", GetMapRoot()),
		fmt.Sprintf("TC_BPF_MNT=%s", GetMapRoot()),
	)
}
