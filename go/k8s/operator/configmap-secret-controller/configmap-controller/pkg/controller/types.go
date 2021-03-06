package controller

import (
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Resource data
type Resource = runtime.Object

// ObjectMeta data
type ObjectMeta = metav1.ObjectMeta

// Pod data
type Pod = v1.Pod

// PodSpec data
type PodSpec = v1.PodSpec

// PodStatus data
type PodStatus = v1.PodStatus

// Node data
type Node = v1.Node

// Namespace data
type Namespace = v1.Namespace

// Container data
type Container = v1.Container

// ContainerPort data
type ContainerPort = v1.ContainerPort

// Event data
type Event = v1.Event

// PodContainerStatus data
type PodContainerStatus = v1.ContainerStatus

// Deployment data
type Deployment = appsv1.Deployment

// ReplicaSet data
type ReplicaSet = appsv1.ReplicaSet

// StatefulSet data
type StatefulSet = appsv1.StatefulSet

// Service data
type Service = v1.Service

type ConfigMap = v1.ConfigMap

type Secret = v1.Secret

const (
	// PodPending phase
	PodPending = v1.PodPending
	// PodRunning phase
	PodRunning = v1.PodRunning
	// PodSucceeded phase
	PodSucceeded = v1.PodSucceeded
	// PodFailed phase
	PodFailed = v1.PodFailed
	// PodUnknown phase
	PodUnknown = v1.PodUnknown
)

// Time extracts time from k8s.Time type
func Time(t *metav1.Time) time.Time {
	return t.Time
}

// ContainerID parses the container ID to get the actual ID string
func ContainerID(s PodContainerStatus) string {
	cID, _ := ContainerIDWithRuntime(s)
	return cID
}

// ContainerIDWithRuntime parses the container ID to get the actual ID string
func ContainerIDWithRuntime(s PodContainerStatus) (string, string) {
	cID := s.ContainerID
	if cID != "" {
		parts := strings.Split(cID, "://")
		if len(parts) == 2 {
			return parts[1], parts[0]
		}
	}
	return "", ""
}
