package storage

import (
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
)

type StorageDriver interface {
	AddStats(cInfo *v1.ContainerInfo, stats *v1.ContainerStats) error

	// Close will clear the state of the storage driver. The elements
	// stored in the underlying storage may or may not be deleted depending
	// on the implementation of the storage driver.
	Close() error
}
