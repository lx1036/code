package k8s

import (
	"github.com/cilium/cilium/pkg/lock"
	"sync"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
)

// CacheAction is the type of action that was performed on the cache
type CacheAction int

const (
	// UpdateService reflects that the service was updated or added
	UpdateService CacheAction = iota

	// DeleteService reflects that the service was deleted
	DeleteService
)

type ServiceEvent struct {
	Action CacheAction
}

type ServiceCache struct {
	mutex sync.RWMutex

	Events chan ServiceEvent
}

func NewServiceCache(nodeAddressing datapath.NodeAddressing) ServiceCache {
	return ServiceCache{

		Events: make(chan ServiceEvent, option.Config.K8sServiceCacheSize),
	}
}

func (s *ServiceCache) UpdateService(k8sSvc *corev1.Service, swg *lock.StoppableWaitGroup) ServiceID {
	svcID, newService := ParseService(k8sSvc, s.nodeAddressing)
	if newService == nil {
		return svcID
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	oldService, ok := s.services[svcID]
	if ok {
		if oldService.DeepEquals(newService) {
			return svcID
		}
	}
	s.services[svcID] = newService

	// Check if the corresponding Endpoints resource is already available
	endpoints, serviceReady := s.correlateEndpoints(svcID)
	if serviceReady {
		swg.Add()
		s.Events <- ServiceEvent{
			Action:     UpdateService,
			ID:         svcID,
			Service:    newService,
			OldService: oldService,
			Endpoints:  endpoints,
			SWG:        swg,
		}
	}

	return svcID
}

func (s *ServiceCache) DeleteService(k8sSvc *corev1.Service, swg *lock.StoppableWaitGroup) {
	svcID := ParseServiceID(k8sSvc)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	oldService, serviceOK := s.services[svcID]
	endpoints, _ := s.correlateEndpoints(svcID)
	delete(s.services, svcID)

	if serviceOK {
		swg.Add()
		s.Events <- ServiceEvent{
			Action:    DeleteService,
			ID:        svcID,
			Service:   oldService,
			Endpoints: endpoints,
			SWG:       swg,
		}
	}
}

func (s *ServiceCache) UpdateEndpoints(k8sEndpoints *corev1.Endpoints, swg *lock.StoppableWaitGroup) (ServiceID, *Endpoints) {
	svcID, newEndpoints := ParseEndpoints(k8sEndpoints)
	epSliceID := EndpointSliceID{
		ServiceID:         svcID,
		EndpointSliceName: k8sEndpoints.GetName(),
	}

	return s.updateEndpoints(epSliceID, newEndpoints, swg)
}

func (s *ServiceCache) UpdateEndpointSlices(epSlice *discoveryv1.EndpointSlice, swg *lock.StoppableWaitGroup) (ServiceID, *Endpoints) {
	svcID, newEndpoints := ParseEndpointSlice(epSlice)

	return s.updateEndpoints(svcID, newEndpoints, swg)
}

func (s *ServiceCache) updateEndpoints(esID EndpointSliceID, newEndpoints *Endpoints, swg *lock.StoppableWaitGroup) (ServiceID, *Endpoints) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if the corresponding Endpoints resource is already available
	svc, ok := s.services[esID.ServiceID]
	endpoints, serviceReady := s.correlateEndpoints(esID.ServiceID)
	if ok && serviceReady {
		swg.Add()
		s.Events <- ServiceEvent{
			Action:    UpdateService,
			ID:        esID.ServiceID,
			Service:   svc,
			Endpoints: endpoints,
			SWG:       swg,
		}
	}

	return esID.ServiceID, newEndpoints
}
