package status

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// PodStatusProvider knows how to provide status for a pod. It's intended to be used by other components
// that need to introspect status.
type PodStatusProvider interface {
	// GetPodStatus returns the cached status for the provided pod UID, as well as whether it
	// was a cache hit.
	GetPodStatus(uid types.UID) (v1.PodStatus, bool)
}
