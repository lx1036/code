package container

import "time"

// RuntimeCache is in interface for obtaining cached Pods.
type RuntimeCache interface {
	GetPods() ([]*Pod, error)
	ForceUpdateIfOlder(time.Time) error
}
