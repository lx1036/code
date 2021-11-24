package controller

import (
	"fmt"
	"k8s.io/klog/v2"
	"reflect"

	"k8s-lx1036/k8s/network/cilium/metallb/pkg/allocator"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/config"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/k8s/types"

	corev1 "k8s.io/api/core/v1"
)

// Service offers methods to mutate a Kubernetes service object.
type service interface {
	UpdateStatus(svc *corev1.Service) error
	Infof(svc *corev1.Service, desc, msg string, args ...interface{})
	Errorf(svc *corev1.Service, desc, msg string, args ...interface{})
}

type Controller struct {
	Client service
	IPs    *allocator.Allocator

	config *config.Config
}

func (c *Controller) SetConfig(cfg *config.Config) types.SyncState {
	if cfg == nil {
		return types.SyncStateError
	}

	if err := c.IPs.SetPools(cfg.Pools); err != nil {
		return types.SyncStateError
	}
	c.config = cfg

	return types.SyncStateReprocessAll
}

func (c *Controller) SetBalancer(key string, rawSvc *corev1.Service, _ *corev1.Endpoints) types.SyncState {
	if rawSvc == nil {
		c.deleteBalancer(key)
		// There might be other LBs stuck waiting for an IP, so when
		// we delete a balancer we should reprocess all of them to
		// check for newly feasible balancers.
		return types.SyncStateReprocessAll
	}
	if c.config == nil {
		// Config hasn't been read, nothing we can do just yet.
		return types.SyncStateSuccess
	}

	// Making a copy unconditionally is a bit wasteful, since we don't
	// always need to update the service. But, making an unconditional
	// copy makes the code much easier to follow, and we have a GC for
	// a reason.
	svc := rawSvc.DeepCopy()
	if !c.allocateService(key, svc) {
		return types.SyncStateError
	}
	if reflect.DeepEqual(rawSvc, svc) {
		klog.Infof(fmt.Sprintf("service %s/%s no change", svc.Namespace, svc.Name))
		return types.SyncStateSuccess
	}

	if !reflect.DeepEqual(rawSvc.Status, svc.Status) {
		// svc 被 allocateService() 后可能不仅仅 status 发生了修改，所以重新 DeepCopy
		updatedSvc := rawSvc.DeepCopy()
		updatedSvc.Status = svc.Status
		if err := c.Client.UpdateStatus(updatedSvc); err != nil {
			klog.Errorf(fmt.Sprintf("failed to update service status: %v", err))
			return types.SyncStateError
		}
	}

	klog.Infof(fmt.Sprintf("allocate loadbalancer ip %s for service %s/%s",
		svc.Status.LoadBalancer.Ingress[0].IP, svc.Namespace, svc.Name))
	return types.SyncStateSuccess
}

func (c *Controller) deleteBalancer(key string) {
	if c.IPs.Unassign(key) {
		klog.Infof(fmt.Sprintf("service deleted"))
	}
}
