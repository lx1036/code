package pod

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// MirrorClient knows how to create/delete a mirror pod in the API server.
type MirrorClient interface {
	// CreateMirrorPod creates a mirror pod in the API server for the given
	// pod or returns an error.  The mirror pod will have the same annotations
	// as the given pod as well as an extra annotation containing the hash of
	// the static pod.
	CreateMirrorPod(pod *v1.Pod) error
	// DeleteMirrorPod deletes the mirror pod with the given full name from
	// the API server or returns an error.
	DeleteMirrorPod(podFullName string, uid *types.UID) (bool, error)
}
