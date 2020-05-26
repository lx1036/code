package common

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

var ListEverything = metav1.ListOptions{
	LabelSelector: labels.Everything().String(),
	FieldSelector: fields.Everything().String(),
}

type ResourceStatus struct {
	// Number of resources that are currently in running state.
	Running int `json:"running"`
	
	// Number of resources that are currently in pending state.
	Pending int `json:"pending"`
	
	// Number of resources that are in failed state.
	Failed int `json:"failed"`
	
	// Number of resources that are in succeeded state.
	Succeeded int `json:"succeeded"`
}

// ListMeta describes list of objects, i.e. holds information about pagination options set for the list.
type ListMeta struct {
	// Total number of items on the list. Used for pagination.
	TotalItems int `json:"totalItems"`
}


