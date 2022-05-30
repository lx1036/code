package controller

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
	"net"
	"reflect"
	"sort"
)

func (controller *NetworkServiceController) onEndpointUpdate(endpoint *corev1.Endpoints) {
	controller.Lock()
	defer controller.Unlock()

	if !controller.readyForUpdates {
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(endpoint)
	if err != nil {
		return
	}
	svc, exist, err := controller.svcLister.GetByKey(key)
	if err != nil {
		klog.Errorf(fmt.Sprintf("failed to get svc %s err: %v", key, err))
		return
	}
	if !exist { // ignore endpoint has no service, if service is deleted, syncService handle it.
		return
	}
	if IsHeadlessService(svc.(*corev1.Service)) || IsExternalNameService(svc.(*corev1.Service)) {
		return
	}

	newSvcMap := controller.buildSvcInfo()
	newEndpointMap := controller.buildEndpointInfo()
	if !newEndpointMap.equal(controller.endpointMap) {
		controller.serviceMap = newSvcMap
		controller.endpointMap = newEndpointMap
		controller.sync()
	} else {
		klog.Infof(fmt.Sprintf("skipping syncing IPVS services for update to endpoint %s as nothing changed", key))
	}
}

type endpointInfo struct {
	address net.IP
	port    int
	isLocal bool
}

type endpointInfoMap map[string][]endpointInfo

func (ep endpointInfoMap) equal(other endpointInfoMap) bool {
	if len(ep) != len(other) {
		return false
	}

	for epID, infos := range ep {
		otherInfos, ok := other[epID]
		if !ok || len(otherInfos) != len(infos) {
			return false
		}
		sort.SliceStable(infos, func(i, j int) bool {
			return infos[i].port < infos[j].port
		})
		sort.SliceStable(otherInfos, func(i, j int) bool {
			return otherInfos[i].port < otherInfos[j].port
		})
		if !reflect.DeepEqual(infos, otherInfos) {
			return false
		}
	}

	return true
}

func (controller *NetworkServiceController) buildEndpointInfo() endpointInfoMap {
	endpointsMap := make(endpointInfoMap)
	for _, obj := range controller.epLister.List() {
		ep := obj.(*corev1.Endpoints)
		for _, subset := range ep.Subsets {
			for _, port := range subset.Ports {
				epID := generateServiceID(ep.Namespace, ep.Name, port.Name)
				var endpoints []endpointInfo
				for _, address := range subset.Addresses {
					endpoints = append(endpoints, endpointInfo{
						address: net.ParseIP(address.IP),
						port:    int(port.Port),
						isLocal: address.NodeName != nil && *address.NodeName == controller.nodeName,
					})
				}

				endpointsMap[epID] = endpoints
			}
		}
	}

	return endpointsMap
}

func hasLocalEndpoint(endpoints []endpointInfo) bool {
	for _, endpoint := range endpoints {
		if endpoint.isLocal {
			return true
		}
	}

	return false
}

func isEndpointsForLeaderElection(ep *corev1.Endpoints) bool {
	_, isLeaderElection := ep.Annotations[resourcelock.LeaderElectionRecordAnnotationKey]
	return isLeaderElection
}
