package lifecycle

import v1 "k8s.io/api/core/v1"

// PodAdmitResult provides the result of a pod admission decision.
type PodAdmitResult struct {
	// if true, the pod should be admitted.
	Admit bool
	// a brief single-word reason why the pod could not be admitted.
	Reason string
	// a brief message explaining why the pod could not be admitted.
	Message string
}

// PodAdmitAttributes is the context for a pod admission decision.
// The member fields of this struct should never be mutated.
type PodAdmitAttributes struct {
	// the pod to evaluate for admission
	Pod *v1.Pod
	// all pods bound to the kubelet excluding the pod being evaluated
	OtherPods []*v1.Pod
}

// PodAdmitHandler is notified during pod admission.
type PodAdmitHandler interface {
	// Admit evaluates if a pod can be admitted.
	Admit(attrs *PodAdmitAttributes) PodAdmitResult
}
