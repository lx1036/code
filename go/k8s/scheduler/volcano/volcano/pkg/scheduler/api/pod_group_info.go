package api

import "volcano.sh/apis/pkg/apis/scheduling"

// PodGroup is a collection of Pod; used for batch workload.
type PodGroup struct {
	scheduling.PodGroup

	// Version represents the version of PodGroup
	Version string
}
