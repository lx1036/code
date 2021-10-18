//go:build linux
// +build linux

package network_namespace

import "syscall"

// file descriptor
type NetworkNamespace int

func None() *NetworkNamespace {
	var netns = NetworkNamespace(-1)
	return &netns
}

func (netns *NetworkNamespace) Close() error {
	if err := syscall.Close(int(*netns)); err != nil {
		return err
	}

	*netns = -1 // clean up

	return nil
}

// get a network namespace from specified path
func GetFromPath(path string) (*NetworkNamespace, error) {
	none := NetworkNamespace(-1)
	fd, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		return &none, err
	}

	netns := NetworkNamespace(fd)
	return &netns, err
}
