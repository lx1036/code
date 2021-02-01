package controller

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

// VolumeStore is an interface that's used to save PersistentVolumes to API server.
// Implementation of the interface add custom error recovery policy.
// A volume is added via StoreVolume(). It's enough to store the volume only once.
// It is not possible to remove a volume, even when corresponding PVC is deleted
// and PV is not necessary any longer. PV will be always created.
// If corresponding PVC is deleted, the PV will be deleted by Kubernetes using
// standard deletion procedure. It saves us some code here.
type VolumeStore interface {
	// StoreVolume makes sure a volume is saved to Kubernetes API server.
	// If no error is returned, caller can assume that PV was saved or
	// is being saved in background.
	// In error is returned, no PV was saved and corresponding PVC needs
	// to be re-queued (so whole provisioning needs to be done again).
	StoreVolume(claim *v1.PersistentVolumeClaim, volume *v1.PersistentVolume) error

	// Runs any background goroutines for implementation of the interface.
	Run(ctx context.Context, threadiness int)
}
