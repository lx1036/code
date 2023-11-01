package service

import (
	"fmt"
	"github.com/cilium/cilium/pkg/counter"
	"github.com/cilium/cilium/pkg/lock"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/maps/lbmap"
	"github.com/cilium/cilium/pkg/option"
	"github.com/sirupsen/logrus"
	nodeTypes "k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/node/types"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/loadbalancer"
	"sync/atomic"
)

// ServiceID is the service's ID.
type ServiceID uint16

// BackendID is the backend's ID.
type BackendID uint16

// ID is the ID of L3n4Addr endpoint (either service or backend).
type ID uint32

// Service is a service handler. Its main responsibility is to reflect
// service-related changes into BPF maps used by datapath BPF programs.
// The changes can be triggered either by k8s_watcher or directly by
// API calls to the /services endpoint.
// 注意：可以被 k8s_watcher 调用，或者直接被 cli /services 调用
type Service struct {
	lock.RWMutex

	svcByHash map[string]*svcInfo
	svcByID   map[loadbalancer.ID]*svcInfo

	backendRefCount counter.StringCounter
	backendByHash   map[string]*loadbalancer.Backend

	healthServer  healthServer
	monitorNotify monitorNotify

	lbmap         LBMap
	lastUpdatedTs atomic.Value
}

// UpsertService inserts or updates the given service.
//
// The first return value is true if the service hasn't existed before.
func (s *Service) UpsertService(params *loadbalancer.SVC) (bool, loadbalancer.ID, error) {
	s.Lock()
	defer s.Unlock()

	scopedLog := log.WithFields(logrus.Fields{
		logfields.ServiceIP: params.Frontend.L3n4Addr,
		logfields.Backends:  params.Backends,

		logfields.ServiceType:                params.Type,
		logfields.ServiceTrafficPolicy:       params.TrafficPolicy,
		logfields.ServiceHealthCheckNodePort: params.HealthCheckNodePort,
		logfields.ServiceName:                params.Name,
		logfields.ServiceNamespace:           params.Namespace,

		logfields.SessionAffinity:        params.SessionAffinity,
		logfields.SessionAffinityTimeout: params.SessionAffinityTimeoutSec,

		logfields.LoadBalancerSourceRanges: params.LoadBalancerSourceRanges,
	})
	scopedLog.Debug("Upserting service")

	if !option.Config.EnableSVCSourceRangeCheck &&
		len(params.LoadBalancerSourceRanges) != 0 {
		scopedLog.Warnf("--%s is disabled, ignoring loadBalancerSourceRanges",
			option.EnableSVCSourceRangeCheck)
	}

	ipv6Svc := params.Frontend.IsIPv6()
	if ipv6Svc && !option.Config.EnableIPv6 {
		err := fmt.Errorf("Unable to upsert service %s as IPv6 is disabled", params.Frontend.L3n4Addr.String())
		return false, lb.ID(0), err
	}
	if !ipv6Svc && !option.Config.EnableIPv4 {
		err := fmt.Errorf("Unable to upsert service %s as IPv4 is disabled", params.Frontend.L3n4Addr.String())
		return false, lb.ID(0), err
	}

	// If needed, create svcInfo and allocate service ID
	svc, new, prevSessionAffinity, prevLoadBalancerSourceRanges, err :=
		s.createSVCInfoIfNotExist(params)
	if err != nil {
		return false, lb.ID(0), err
	}
	// TODO(brb) defer ServiceID release after we have a lbmap "rollback"
	scopedLog = scopedLog.WithField(logfields.ServiceID, svc.frontend.ID)
	scopedLog.Debug("Acquired service ID")

	onlyLocalBackends, filterBackends := svc.requireNodeLocalBackends(params.Frontend)
	prevBackendCount := len(svc.backends)

	backendsCopy := []lb.Backend{}
	for _, b := range params.Backends {
		// Local redirect services or services with trafficPolicy=Local may
		// only use node-local backends for external scope. We implement this by
		// filtering out all backend IPs which are not a local endpoint.
		if filterBackends && len(b.NodeName) > 0 && b.NodeName != nodeTypes.GetName() {
			continue
		}
		backendsCopy = append(backendsCopy, *b.DeepCopy())
	}

	// TODO (Aditi) When we filter backends for LocalRedirect service, there
	// might be some backend pods with active connections. We may need to
	// defer filtering the backends list (thereby defer redirecting traffic)
	// in such cases. GH #12859
	// Update backends cache and allocate/release backend IDs
	newBackends, obsoleteBackendIDs, obsoleteSVCBackendIDs, err :=
		s.updateBackendsCacheLocked(svc, backendsCopy)
	if err != nil {
		return false, lb.ID(0), err
	}

	// Update lbmaps (BPF service maps)
	if err = s.upsertServiceIntoLBMaps(svc, onlyLocalBackends, prevBackendCount, newBackends,
		obsoleteBackendIDs, prevSessionAffinity,
		prevLoadBalancerSourceRanges, obsoleteSVCBackendIDs,
		scopedLog); err != nil {

		return false, lb.ID(0), err
	}

	// Only add a HealthCheckNodePort server if this is a service which may
	// only contain local backends (i.e. it has externalTrafficPolicy=Local)
	if option.Config.EnableHealthCheckNodePort {
		if onlyLocalBackends && filterBackends {
			localBackendCount := len(backendsCopy)
			s.healthServer.UpsertService(lb.ID(svc.frontend.ID), svc.svcNamespace, svc.svcName,
				localBackendCount, svc.svcHealthCheckNodePort)
		} else if svc.svcHealthCheckNodePort == 0 {
			// Remove the health check server in case this service used to have
			// externalTrafficPolicy=Local with HealthCheckNodePort in the previous
			// version, but not anymore.
			s.healthServer.DeleteService(lb.ID(svc.frontend.ID))
		}
	}

	if new {
		addMetric.Inc()
	} else {
		updateMetric.Inc()
	}

	s.notifyMonitorServiceUpsert(svc.frontend, svc.backends,
		svc.svcType, svc.svcTrafficPolicy, svc.svcName, svc.svcNamespace)
	return new, lb.ID(svc.frontend.ID), nil
}

func (s *Service) upsertServiceIntoLBMaps(svc *svcInfo, onlyLocalBackends bool,
	prevBackendCount int, newBackends []lb.Backend, obsoleteBackendIDs []lb.BackendID,
	prevSessionAffinity bool, prevLoadBalancerSourceRanges []*cidr.CIDR,
	obsoleteSVCBackendIDs []lb.BackendID, scopedLog *logrus.Entry) error {

	ipv6 := svc.frontend.IsIPv6()

	var (
		toDeleteAffinity, toAddAffinity []lb.BackendID
		checkLBSrcRange                 bool
	)

	// Update sessionAffinity
	if option.Config.EnableSessionAffinity {
		if prevSessionAffinity && !svc.sessionAffinity {
			// Remove backends from the affinity match because the svc's sessionAffinity
			// has been disabled
			toDeleteAffinity = make([]lb.BackendID, 0, len(obsoleteSVCBackendIDs)+len(svc.backends))
			toDeleteAffinity = append(toDeleteAffinity, obsoleteSVCBackendIDs...)
			for _, b := range svc.backends {
				toDeleteAffinity = append(toDeleteAffinity, b.ID)
			}
		} else if svc.sessionAffinity {
			toAddAffinity = make([]lb.BackendID, 0, len(svc.backends))
			for _, b := range svc.backends {
				toAddAffinity = append(toAddAffinity, b.ID)
			}
			if prevSessionAffinity {
				// Remove obsolete svc backends if previously the svc had the affinity enabled
				toDeleteAffinity = make([]lb.BackendID, 0, len(obsoleteSVCBackendIDs))
				for _, bID := range obsoleteSVCBackendIDs {
					toDeleteAffinity = append(toDeleteAffinity, bID)
				}
			}
		}

		s.deleteBackendsFromAffinityMatchMap(svc.frontend.ID, toDeleteAffinity)
		// New affinity matches (toAddAffinity) will be added after the new
		// backends have been added.
	}

	// Update LB source range check cidrs
	if option.Config.EnableSVCSourceRangeCheck {
		checkLBSrcRange = len(svc.loadBalancerSourceRanges) != 0
		if checkLBSrcRange || len(prevLoadBalancerSourceRanges) != 0 {
			if err := s.lbmap.UpdateSourceRanges(uint16(svc.frontend.ID),
				prevLoadBalancerSourceRanges, svc.loadBalancerSourceRanges,
				ipv6); err != nil {

				return err
			}
		}
	}

	// Add new backends into BPF maps
	for _, b := range newBackends {
		scopedLog.WithFields(logrus.Fields{
			logfields.BackendID: b.ID,
			logfields.L3n4Addr:  b.L3n4Addr,
		}).Debug("Adding new backend")

		if err := s.lbmap.AddBackend(uint16(b.ID), b.L3n4Addr.IP,
			b.L3n4Addr.L4Addr.Port, ipv6); err != nil {
			return err
		}
	}

	// Upsert service entries into BPF maps
	backends := make(map[string]uint16, len(svc.backends))
	for _, b := range svc.backends {
		backends[b.String()] = uint16(b.ID)
	}

	p := &lbmap.UpsertServiceParams{
		ID:                        uint16(svc.frontend.ID),
		IP:                        svc.frontend.L3n4Addr.IP,
		Port:                      svc.frontend.L3n4Addr.L4Addr.Port,
		Backends:                  backends,
		PrevBackendCount:          prevBackendCount,
		IPv6:                      ipv6,
		Type:                      svc.svcType,
		Local:                     onlyLocalBackends,
		Scope:                     svc.frontend.L3n4Addr.Scope,
		SessionAffinity:           svc.sessionAffinity,
		SessionAffinityTimeoutSec: svc.sessionAffinityTimeoutSec,
		CheckSourceRange:          checkLBSrcRange,
		UseMaglev:                 svc.useMaglev(),
	}
	if err := s.lbmap.UpsertService(p); err != nil {
		return err
	}

	if option.Config.EnableSessionAffinity {
		s.addBackendsToAffinityMatchMap(svc.frontend.ID, toAddAffinity)
	}

	// Remove backends not used by any service from BPF maps
	for _, id := range obsoleteBackendIDs {
		scopedLog.WithField(logfields.BackendID, id).
			Debug("Removing obsolete backend")

		if err := s.lbmap.DeleteBackendByID(uint16(id), ipv6); err != nil {
			log.WithError(err).WithField(logfields.BackendID, id).
				Warn("Failed to remove backend from maps")
		}
	}

	return nil
}
