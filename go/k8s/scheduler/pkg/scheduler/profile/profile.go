package profile

import (
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	"k8s.io/client-go/tools/events"
)

// Profile is a scheduling profile.
type Profile struct {
	v1alpha1.Framework
	Recorder events.EventRecorder
	Name     string
}
