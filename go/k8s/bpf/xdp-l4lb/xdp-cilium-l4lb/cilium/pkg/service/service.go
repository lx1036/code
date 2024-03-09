package service

import (
	"fmt"
	"github.com/cilium/cilium/pkg/counter"
	"github.com/cilium/cilium/pkg/lock"
	"github.com/sirupsen/logrus"
	nodeTypes "k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/node/types"
	"net"
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/cidr"
	datapathOption "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/option"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/loadbalancer"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/lbmap"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/service/healthserver"
	"sync/atomic"
)

var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "service")

// ServiceID is the service's ID.
type ServiceID uint16

// BackendID is the backend's ID.
type BackendID uint16

// ID is the ID of L3n4Addr endpoint (either service or backend).
type ID uint32

// Service 注意：可以被 k8s_watcher 调用，或者直接被 cli /services 调用
type Service struct {
	lock.RWMutex

	svcByHash map[string]*svcInfo
	svcByID   map[loadbalancer.ID]*svcInfo

	backendRefCount counter.StringCounter
	backendByHash   map[string]*loadbalancer.Backend

	healthServer  healthServer
	monitorNotify monitorNotify

	lbmap         *lbmap.LBBPFMap
	lastUpdatedTs atomic.Value
}

// healthServer is used to manage HealtCheckNodePort listeners
type healthServer interface {
	UpsertService(svcID loadbalancer.ID, svcNS, svcName string, localEndpoints int, port uint16)
	DeleteService(svcID loadbalancer.ID)
}

func NewService(monitorNotify monitorNotify) *Service {
	var localHealthServer healthServer
	if option.Config.EnableHealthCheckNodePort {
		localHealthServer = healthserver.New()
	}

	maglev := option.Config.NodePortAlg == option.NodePortAlgMaglev
	maglevTableSize := option.Config.MaglevTableSize

	svc := &Service{
		svcByHash:       map[string]*svcInfo{},
		svcByID:         map[loadbalancer.ID]*svcInfo{},
		backendRefCount: counter.StringCounter{},
		backendByHash:   map[string]*loadbalancer.Backend{},
		monitorNotify:   monitorNotify,
		healthServer:    localHealthServer,
		lbmap:           lbmap.New(maglev, maglevTableSize),
	}
	svc.lastUpdatedTs.Store(time.Now())
	return svc
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
		return false, loadbalancer.ID(0), err
	}
	if !ipv6Svc && !option.Config.EnableIPv4 {
		err := fmt.Errorf("Unable to upsert service %s as IPv4 is disabled", params.Frontend.L3n4Addr.String())
		return false, loadbalancer.ID(0), err
	}

	// If needed, create svcInfo and allocate service ID
	svc, n, prevSessionAffinity, prevLoadBalancerSourceRanges, err := s.createSVCInfoIfNotExist(params)
	if err != nil {
		return false, loadbalancer.ID(0), err
	}
	// TODO(brb) defer ServiceID release after we have a lbmap "rollback"
	scopedLog = scopedLog.WithField(logfields.ServiceID, svc.frontend.ID)
	scopedLog.Debug("Acquired service ID")

	onlyLocalBackends, filterBackends := svc.requireNodeLocalBackends(params.Frontend)
	prevBackendCount := len(svc.backends)

	backendsCopy := []loadbalancer.Backend{}
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
	newBackends, obsoleteBackendIDs, obsoleteSVCBackendIDs, err := s.updateBackendsCacheLocked(svc, backendsCopy)
	if err != nil {
		return false, loadbalancer.ID(0), err
	}

	// Update lbmaps (BPF service maps)
	if err = s.upsertServiceIntoLBMaps(svc, onlyLocalBackends, prevBackendCount, newBackends,
		obsoleteBackendIDs, prevSessionAffinity,
		prevLoadBalancerSourceRanges, obsoleteSVCBackendIDs,
		scopedLog); err != nil {

		return false, loadbalancer.ID(0), err
	}

	// Only add a HealthCheckNodePort server if this is a service which may
	// only contain local backends (i.e. it has externalTrafficPolicy=Local)
	if option.Config.EnableHealthCheckNodePort {
		if onlyLocalBackends && filterBackends {
			localBackendCount := len(backendsCopy)
			s.healthServer.UpsertService(loadbalancer.ID(svc.frontend.ID), svc.svcNamespace, svc.svcName,
				localBackendCount, svc.svcHealthCheckNodePort)
		} else if svc.svcHealthCheckNodePort == 0 {
			// Remove the health check server in case this service used to have
			// externalTrafficPolicy=Local with HealthCheckNodePort in the previous
			// version, but not anymore.
			s.healthServer.DeleteService(loadbalancer.ID(svc.frontend.ID))
		}
	}

	if n {
		addMetric.Inc()
	} else {
		updateMetric.Inc()
	}

	s.notifyMonitorServiceUpsert(svc.frontend, svc.backends,
		svc.svcType, svc.svcTrafficPolicy, svc.svcName, svc.svcNamespace)
	return n, loadbalancer.ID(svc.frontend.ID), nil
}

func (s *Service) upsertServiceIntoLBMaps(svc *svcInfo, onlyLocalBackends bool,
	prevBackendCount int, newBackends []loadbalancer.Backend, obsoleteBackendIDs []loadbalancer.BackendID,
	prevSessionAffinity bool, prevLoadBalancerSourceRanges []*cidr.CIDR,
	obsoleteSVCBackendIDs []loadbalancer.BackendID, scopedLog *logrus.Entry) error {

	ipv6 := svc.frontend.IsIPv6()

	var (
		toDeleteAffinity, toAddAffinity []loadbalancer.BackendID
		checkLBSrcRange                 bool
	)

	// Update sessionAffinity
	if option.Config.EnableSessionAffinity {
		if prevSessionAffinity && !svc.sessionAffinity {
			// Remove backends from the affinity match because the svc's sessionAffinity
			// has been disabled
			toDeleteAffinity = make([]loadbalancer.BackendID, 0, len(obsoleteSVCBackendIDs)+len(svc.backends))
			toDeleteAffinity = append(toDeleteAffinity, obsoleteSVCBackendIDs...)
			for _, b := range svc.backends {
				toDeleteAffinity = append(toDeleteAffinity, b.ID)
			}
		} else if svc.sessionAffinity {
			toAddAffinity = make([]loadbalancer.BackendID, 0, len(svc.backends))
			for _, b := range svc.backends {
				toAddAffinity = append(toAddAffinity, b.ID)
			}
			if prevSessionAffinity {
				// Remove obsolete svc backends if previously the svc had the affinity enabled
				toDeleteAffinity = make([]loadbalancer.BackendID, 0, len(obsoleteSVCBackendIDs))
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

// RestoreServices restores services from BPF maps.
func (s *Service) RestoreServices() error {
	s.Lock()
	defer s.Unlock()

	// Restore backend IDs
	if err := s.restoreBackendsLocked(); err != nil {
		return err
	}

	// Restore service cache from BPF maps
	if err := s.restoreServicesLocked(); err != nil {
		return err
	}

	// Remove no longer existing affinity matches
	if option.Config.EnableSessionAffinity {
		if err := s.deleteOrphanAffinityMatchesLocked(); err != nil {
			return err
		}
	}

	// Remove LB source ranges for no longer existing services
	if option.Config.EnableSVCSourceRangeCheck {
		if err := s.restoreAndDeleteOrphanSourceRanges(); err != nil {
			return err
		}
	}

	// Remove obsolete backends and release their IDs
	if err := s.deleteOrphanBackends(); err != nil {
		log.WithError(err).Warn("Failed to remove orphan backends")

	}

	return nil
}

type svcInfo struct {
	hash          string
	frontend      loadbalancer.L3n4AddrID
	backends      []loadbalancer.Backend
	backendByHash map[string]*loadbalancer.Backend

	svcType                   loadbalancer.SVCType
	svcTrafficPolicy          loadbalancer.SVCTrafficPolicy
	sessionAffinity           bool
	sessionAffinityTimeoutSec uint32
	svcHealthCheckNodePort    uint16
	svcName                   string
	svcNamespace              string
	loadBalancerSourceRanges  []*cidr.CIDR

	restoredFromDatapath bool
}

func (svc *svcInfo) useMaglev() bool {
	return option.Config.NodePortAlg == option.NodePortAlgMaglev &&
		((svc.svcType == loadbalancer.SVCTypeNodePort && !isWildcardAddr(svc.frontend)) ||
			svc.svcType == loadbalancer.SVCTypeExternalIPs ||
			svc.svcType == loadbalancer.SVCTypeLoadBalancer)
}

func (s *Service) restoreServicesLocked() error {
	svcs, errors := s.lbmap.DumpServiceMaps()
	for _, err := range errors {
		log.WithError(err).Warning("Error occurred while dumping service maps")
	}

	for _, svc := range svcs {
		scopedLog := log.WithFields(logrus.Fields{
			logfields.ServiceID: svc.Frontend.ID,
			logfields.ServiceIP: svc.Frontend.L3n4Addr.String(),
		})
		scopedLog.Debug("Restoring service")

		if _, err := RestoreID(svc.Frontend.L3n4Addr, uint32(svc.Frontend.ID)); err != nil {
			failed++
			scopedLog.WithError(err).Warning("Unable to restore service ID")
		}

		newSVC := &svcInfo{
			hash:             svc.Frontend.Hash(),
			frontend:         svc.Frontend,
			backends:         svc.Backends,
			backendByHash:    map[string]*loadbalancer.Backend{},
			svcType:          svc.Type,
			svcTrafficPolicy: svc.TrafficPolicy,

			sessionAffinity:           svc.SessionAffinity,
			sessionAffinityTimeoutSec: svc.SessionAffinityTimeoutSec,

			// Indicate that the svc was restored from the BPF maps, so that
			// SyncWithK8sFinished() could remove services which were restored
			// from the maps but not present in the k8sServiceCache (e.g. a svc
			// was deleted while cilium-agent was down).
			restoredFromDatapath: true,
		}

		for j, backend := range svc.Backends {
			hash := backend.L3n4Addr.Hash()
			s.backendRefCount.Add(hash)
			newSVC.backendByHash[hash] = &svc.Backends[j]
		}

		// Recalculate Maglev lookup tables if the maps were removed due to the changed M param.
		ipv6 := newSVC.frontend.IsIPv6()
		if option.Config.DatapathMode == datapathOption.DatapathModeLBOnly &&
			newSVC.useMaglev() && s.lbmap.IsMaglevLookupTableRecreated(ipv6) {

			backends := make(map[string]uint16, len(newSVC.backends))
			for _, b := range newSVC.backends {
				backends[b.String()] = uint16(b.ID)
			}
			if err := s.lbmap.UpsertMaglevLookupTable(uint16(newSVC.frontend.ID), backends, ipv6); err != nil {
				return err
			}
		}

		s.svcByHash[newSVC.hash] = newSVC
		s.svcByID[newSVC.frontend.ID] = newSVC
		restored++
	}

	log.WithFields(logrus.Fields{
		"restored": restored,
		"failed":   failed,
	}).Info("Restored services from maps")

	return nil
}

// isWildcardAddr returns true if given frontend is used for wildcard svc lookups
// (by bpf_sock).
func isWildcardAddr(frontend loadbalancer.L3n4AddrID) bool {
	if frontend.IsIPv6() {
		return net.IPv6zero.Equal(frontend.IP)
	}
	return net.IPv4zero.Equal(frontend.IP)
}
