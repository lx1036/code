package app

import (
	"context"
	"github.com/cilium/cilium/pkg/controller"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/endpoint/endpointmanager"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/service"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath/loader"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/watchers"
	"time"
)

type Daemon struct {
	k8sWatcher      *watchers.K8sWatcher
	k8sCachesSynced <-chan struct{}

	serviceBPFManager *service.ServiceBPFManager

	endpointManager *endpointmanager.EndpointManager
}

func NewDaemon() (*Daemon, *endpointRestoreState, error) {

	d := Daemon{
		endpointManager: endpointmanager.NewEndpointManager(&watchers.EndpointSynchronizer{}),
	}
	d.endpointManager.InitMetrics()

	d.serviceBPFManager = service.NewServiceBPFManager(&d)

	d.k8sWatcher = watchers.NewK8sWatcher(d.endpointManager, d.serviceBPFManager)

	// (1) Open or create BPF maps.
	err = d.initMaps()

	// Read the service IDs of existing services from the BPF map and
	// reserve them. This must be done *before* connecting to the
	// Kubernetes apiserver and serving the API to ensure service IDs are
	// not changing across restarts or that a new service could accidentally
	// use an existing service ID.
	// Also, create missing v2 services from the corresponding legacy ones.
	d.serviceBPFManager.RestoreServices()

	d.k8sWatcher.RunK8sServiceHandler()

	d.k8sCachesSynced = d.k8sWatcher.InitK8sSubsystem()

	// restore endpoints before any IPs are allocated to avoid eventual IP
	// conflicts later on, otherwise any IP conflict will result in the
	// endpoint not being able to be restored.
	restoredEndpoints, err := d.restoreOldEndpoints(option.Config.StateDir, true)
	if err != nil {
		log.WithError(err).Error("Unable to restore existing endpoints")
	}

	if err := d.allocateIPs(); err != nil {
		return nil, nil, err
	}
	
	
	err = d.init()
	
	
	if err := d.syncEndpointsAndHostIPs(); err != nil {
		return nil, nil, err
	}
	// Start the controller for periodic sync. The purpose of the
	// controller is to ensure that endpoints and host IPs entries are
	// reinserted to the bpf maps if they are ever removed from them.
	controller.NewManager().UpdateController("sync-endpoints-and-host-ips",
		controller.ControllerParams{
			DoFunc: func(ctx context.Context) error {
				return d.syncEndpointsAndHostIPs()
			},
			RunInterval: time.Minute,
			Context:     d.ctx,
		})

	if err := loader.RestoreTemplates(option.Config.StateDir); err != nil {
		log.WithError(err).Error("Unable to restore previous BPF templates")
	}

	return &d, restoredEndpoints, nil
}
