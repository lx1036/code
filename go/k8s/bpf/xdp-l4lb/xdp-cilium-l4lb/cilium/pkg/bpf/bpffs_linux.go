//go:build linux

package bpf

import (
	"path/filepath"
	"sync"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/defaults"
)

var (
	once sync.Once

	// Set to true on first get request to detect misorder
	lockedDown = false

	mapRoot = defaults.DefaultMapRoot

	// Prefix for all maps (default: tc/globals)
	mapPrefix = defaults.DefaultMapPrefix
)

func lockDown() {
	lockedDown = true
}

func MapPath(name string) string {
	once.Do(lockDown)
	return filepath.Join(mapRoot, mapPrefix, name)
}
