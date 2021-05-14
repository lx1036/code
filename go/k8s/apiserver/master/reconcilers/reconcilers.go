package reconcilers

import (
	"net"

	corev1 "k8s.io/api/core/v1"
)

// EndpointReconciler knows how to reconcile the endpoints for the apiserver service.
type EndpointReconciler interface {
	// ReconcileEndpoints sets the endpoints for the given apiserver service (ro or rw).
	// ReconcileEndpoints expects that the endpoints objects it manages will all be
	// managed only by ReconcileEndpoints; therefore, to understand this, you need only
	// understand the requirements.
	//
	// Requirements:
	//  * All apiservers MUST use the same ports for their {rw, ro} services.
	//  * All apiservers MUST use ReconcileEndpoints and only ReconcileEndpoints to manage the
	//      endpoints for their {rw, ro} services.
	//  * ReconcileEndpoints is called periodically from all apiservers.
	ReconcileEndpoints(serviceName string, ip net.IP, endpointPorts []corev1.EndpointPort, reconcilePorts bool) error
	// RemoveEndpoints removes this apiserver's lease.
	RemoveEndpoints(serviceName string, ip net.IP, endpointPorts []corev1.EndpointPort) error
	// StopReconciling turns any later ReconcileEndpoints call into a noop.
	StopReconciling()
}

// Type the reconciler type
type Type string

const (
	// MasterCountReconcilerType will select the original reconciler
	MasterCountReconcilerType Type = "master-count"
	// LeaseEndpointReconcilerType will select a storage based reconciler
	LeaseEndpointReconcilerType Type = "lease"
	// NoneEndpointReconcilerType will turn off the endpoint reconciler
	NoneEndpointReconcilerType Type = "none"
)
