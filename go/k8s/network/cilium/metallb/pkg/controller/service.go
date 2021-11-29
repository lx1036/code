package controller

import (
	"fmt"
	"k8s.io/klog/v2"
	"net"

	corev1 "k8s.io/api/core/v1"
)

func (c *Controller) allocateService(key string, svc *corev1.Service) bool {
	var lbIP net.IP

	// Not a LoadBalancer, early exit. It might have been a balancer
	// in the past, so we still need to clear LB state.
	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
		c.clearServiceState(key, svc)
		return true
	}

	// If the ClusterIP is malformed or not set we can't determine the
	// ipFamily to use.
	clusterIP := net.ParseIP(svc.Spec.ClusterIP)
	if clusterIP == nil {
		klog.Infof(fmt.Sprintf("No ClusterIP"))
		c.clearServiceState(key, svc)
		return true
	}

	// The assigned LB IP is the end state of convergence. If there's
	// none or a malformed one, nuke all controlled state so that we
	// start converging from a clean slate.
	if len(svc.Status.LoadBalancer.Ingress) == 1 {
		lbIP = net.ParseIP(svc.Status.LoadBalancer.Ingress[0].IP)
	}
	if lbIP == nil {
		c.clearServiceState(key, svc)
	}

	// 要么是 ipv4 要么是 ipv6
	// Clear the lbIP if it has a different ipFamily compared to the clusterIP.
	// (this should not happen since the "ipFamily" of a service is immutable)
	if (clusterIP.To4() == nil) != (lbIP.To4() == nil) {
		c.clearServiceState(key, svc)
		lbIP = nil
	}

	// It's possible the config mutated and the IP we have no longer
	// makes sense. If so, clear it out and give the rest of the logic
	// a chance to allocate again.
	if lbIP != nil {
		// This assign is idempotent if the config is consistent,
		// otherwise it'll fail and tell us why.
		if err := c.IPs.Assign(key, lbIP, Ports(svc), SharingKey(svc), BackendKey(svc)); err != nil {
			//l.Log("event", "clearAssignment", "reason", "notAllowedByConfig", "msg", "current IP not allowed by config, clearing")
			c.clearServiceState(key, svc)
			lbIP = nil
		}

		// The user might also have changed the pool annotation, and
		// requested a different pool than the one that is currently
		// allocated.
		desiredPool := svc.Annotations["metallb.universe.tf/address-pool"]
		if lbIP != nil && desiredPool != "" && c.IPs.Pool(key) != desiredPool {
			c.clearServiceState(key, svc)
			lbIP = nil
		}
	}

	// User set or changed the desired LB IP, nuke the
	// state. allocateIP will pay attention to LoadBalancerIP and try
	// to meet the user's demands.
	if svc.Spec.LoadBalancerIP != "" && svc.Spec.LoadBalancerIP != lbIP.String() {
		c.clearServiceState(key, svc)
		lbIP = nil
	}

	// If lbIP is still nil at this point, try to allocate.
	if lbIP == nil {
		ip, err := c.allocateIP(key, svc)
		if err != nil {
			klog.Errorf(fmt.Sprintf("Failed to allocate IP for %q: %s", key, err))
			return true
		}
		lbIP = ip
		klog.Infof(fmt.Sprintf("Assigned IP %q", lbIP))
	}

	if lbIP == nil {
		c.clearServiceState(key, svc)
		return true
	}

	pool := c.IPs.Pool(key)
	if pool == "" || c.config.Pools[pool] == nil {
		c.clearServiceState(key, svc)
		return true
	}

	// At this point, we have an IP selected somehow, all that remains
	// is to program the data plane.
	svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: lbIP.String()}}
	return true
}

func (c *Controller) allocateIP(key string, svc *corev1.Service) (net.IP, error) {
	clusterIP := net.ParseIP(svc.Spec.ClusterIP)
	if clusterIP == nil {
		// (we should never get here because the caller ensured that Spec.ClusterIP != nil)
		return nil, fmt.Errorf("invalid ClusterIP [%s], can't determine family", svc.Spec.ClusterIP)
	}
	isIPv6 := clusterIP.To4() == nil

	// If the user asked for a specific IP, try that.
	if svc.Spec.LoadBalancerIP != "" {
		ip := net.ParseIP(svc.Spec.LoadBalancerIP)
		if ip == nil {
			return nil, fmt.Errorf("invalid spec.loadBalancerIP %q", svc.Spec.LoadBalancerIP)
		}
		if (ip.To4() == nil) != isIPv6 {
			return nil, fmt.Errorf("requested spec.loadBalancerIP %q does not match the ipFamily of the service", svc.Spec.LoadBalancerIP)
		}
		if err := c.IPs.Assign(key, ip, Ports(svc), SharingKey(svc), BackendKey(svc)); err != nil {
			return nil, err
		}
		return ip, nil
	}

	// Otherwise, did the user ask for a specific pool?
	desiredPool := svc.Annotations["metallb.universe.tf/address-pool"]
	if desiredPool != "" {
		ip, err := c.IPs.AllocateFromPool(key, isIPv6, desiredPool, Ports(svc), SharingKey(svc), BackendKey(svc))
		if err != nil {
			return nil, err
		}
		return ip, nil
	}

	// Okay, in that case just bruteforce across all pools.
	return c.IPs.Allocate(key, isIPv6, Ports(svc), SharingKey(svc), BackendKey(svc))
}

// clearServiceState clears all fields that are actively managed by
// this controller.
func (c *Controller) clearServiceState(key string, svc *corev1.Service) {
	c.IPs.Unassign(key)
	svc.Status.LoadBalancer = corev1.LoadBalancerStatus{}
}
