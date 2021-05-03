package pod

import (
	kubetypes "k8s-lx1036/k8s/kubelet/pkg/types"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Manager stores and manages access to pods, maintaining the mappings
// between static pods and mirror pods.
//
// The kubelet discovers pod updates from 3 sources: file, http, and
// apiserver. Pods from non-apiserver sources are called static pods, and API
// server is not aware of the existence of static pods. In order to monitor
// the status of such pods, the kubelet creates a mirror pod for each static
// pod via the API server.
//
// A mirror pod has the same pod full name (name and namespace) as its static
// counterpart (albeit different metadata such as UID, etc). By leveraging the
// fact that the kubelet reports the pod status using the pod full name, the
// status of the mirror pod always reflects the actual status of the static
// pod. When a static pod gets deleted, the associated orphaned mirror pod
// will also be removed.
type Manager interface {
	// GetPods returns the regular pods bound to the kubelet and their spec.
	GetPods() []*v1.Pod
	// GetPodByFullName returns the (non-mirror) pod that matches full name, as well as
	// whether the pod was found.
	GetPodByFullName(podFullName string) (*v1.Pod, bool)
	// GetPodByName provides the (non-mirror) pod that matches namespace and
	// name, as well as whether the pod was found.
	GetPodByName(namespace, name string) (*v1.Pod, bool)
	// GetPodByUID provides the (non-mirror) pod that matches pod UID, as well as
	// whether the pod is found.
	GetPodByUID(types.UID) (*v1.Pod, bool)
	// GetPodByMirrorPod returns the static pod for the given mirror pod and
	// whether it was known to the pod manager.
	GetPodByMirrorPod(*v1.Pod) (*v1.Pod, bool)
	// GetMirrorPodByPod returns the mirror pod for the given static pod and
	// whether it was known to the pod manager.
	GetMirrorPodByPod(*v1.Pod) (*v1.Pod, bool)
	// GetPodsAndMirrorPods returns the both regular and mirror pods.
	GetPodsAndMirrorPods() ([]*v1.Pod, []*v1.Pod)
	// SetPods replaces the internal pods with the new pods.
	// It is currently only used for testing.
	SetPods(pods []*v1.Pod)
	// AddPod adds the given pod to the manager.
	AddPod(pod *v1.Pod)
	// UpdatePod updates the given pod in the manager.
	UpdatePod(pod *v1.Pod)
	// DeletePod deletes the given pod from the manager.  For mirror pods,
	// this means deleting the mappings related to mirror pods.  For non-
	// mirror pods, this means deleting from indexes for all non-mirror pods.
	DeletePod(pod *v1.Pod)
	// GetOrphanedMirrorPodNames returns names of orphaned mirror pods
	GetOrphanedMirrorPodNames() []string
	// TranslatePodUID returns the actual UID of a pod. If the UID belongs to
	// a mirror pod, returns the UID of its static pod. Otherwise, returns the
	// original UID.
	//
	// All public-facing functions should perform this translation for UIDs
	// because user may provide a mirror pod UID, which is not recognized by
	// internal Kubelet functions.
	TranslatePodUID(uid types.UID) kubetypes.ResolvedPodUID
	// GetUIDTranslations returns the mappings of static pod UIDs to mirror pod
	// UIDs and mirror pod UIDs to static pod UIDs.
	GetUIDTranslations() (podToMirror map[kubetypes.ResolvedPodUID]kubetypes.MirrorPodUID, mirrorToPod map[kubetypes.MirrorPodUID]kubetypes.ResolvedPodUID)
	// IsMirrorPodOf returns true if mirrorPod is a correct representation of
	// pod; false otherwise.
	IsMirrorPodOf(mirrorPod, pod *v1.Pod) bool

	MirrorClient
}
