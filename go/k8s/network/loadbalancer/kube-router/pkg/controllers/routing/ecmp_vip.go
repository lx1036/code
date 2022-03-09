package routing

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	gobgpapi "github.com/osrg/gobgp/api"
	bgppacket "github.com/osrg/gobgp/pkg/packet/bgp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// INFO: ECMP(Equal Cost Multi-Path) 等价路由: 多条不同链路到达同一目的地址的网络环境，即同一个 dst 多个 next hop

func getNextHop(path *gobgpapi.Path) string {
	for _, pattr := range path.Pattrs {
		var msg ptypes.DynamicAny
		ptypes.UnmarshalAny(pattr, &msg)
		switch t := msg.Message.(type) {
		case *gobgpapi.NextHopAttribute:
			return t.NextHop
		}
	}

	return ""
}

func (controller *NetworkRoutingController) isVIPExistedInTable(vip string) bool {
	existed := false
	err := controller.bgpServer.ListPath(context.Background(), &gobgpapi.ListPathRequest{
		TableType: gobgpapi.TableType_GLOBAL,
		Family:    &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
		Prefixes: []*gobgpapi.TableLookupPrefix{
			{
				Prefix: vip,
			},
		},
	}, func(destination *gobgpapi.Destination) {
		for _, path := range destination.Paths {
			if getNextHop(path) == controller.nodeIP.String() {
				existed = true
			}
		}
	})

	return err == nil && existed
}

func (controller *NetworkRoutingController) advertiseVIPs(vips []string) {
	for _, vip := range vips {
		klog.Infof(fmt.Sprintf("advertising route: '%s/32 via %s' to peers", vip, controller.nodeIP.String()))

		if controller.isVIPExistedInTable(vip) {
			continue
		}

		a1, _ := ptypes.MarshalAny(&gobgpapi.OriginAttribute{
			Origin: uint32(bgppacket.BGP_ORIGIN_ATTR_TYPE_IGP),
		})
		a2, _ := ptypes.MarshalAny(&gobgpapi.NextHopAttribute{
			NextHop: controller.nodeIP.String(),
		})
		attrs := []*any.Any{a1, a2}
		nlri, _ := ptypes.MarshalAny(&gobgpapi.IPAddressPrefix{
			Prefix:    vip,
			PrefixLen: 32,
		})
		_, err := controller.bgpServer.AddPath(context.Background(), &gobgpapi.AddPathRequest{
			TableType: gobgpapi.TableType_GLOBAL,
			Path: &gobgpapi.Path{
				Family: &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
				Nlri:   nlri,
				Pattrs: attrs,
			},
		})
		if err != nil {
			klog.Errorf(fmt.Sprintf("advertising IP: %q, error: %v", vip, err))
		}
	}
}

func (controller *NetworkRoutingController) withdrawVIPs(vips []string) {
	for _, vip := range vips {
		klog.Infof(fmt.Sprintf("withdrawing route: '%s/32 via %s' to peers", vip, controller.nodeIP.String()))

		if !controller.isVIPExistedInTable(vip) {
			continue
		}

		a1, _ := ptypes.MarshalAny(&gobgpapi.OriginAttribute{
			Origin: 0,
		})
		a2, _ := ptypes.MarshalAny(&gobgpapi.NextHopAttribute{
			NextHop: controller.nodeIP.String(),
		})
		attrs := []*any.Any{a1, a2}
		nlri, _ := ptypes.MarshalAny(&gobgpapi.IPAddressPrefix{
			Prefix:    vip,
			PrefixLen: 32,
		})
		err := controller.bgpServer.DeletePath(context.Background(), &gobgpapi.DeletePathRequest{
			TableType: gobgpapi.TableType_GLOBAL,
			Path: &gobgpapi.Path{
				Family: &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
				Nlri:   nlri,
				Pattrs: attrs,
			},
		})
		if err != nil {
			klog.Errorf(fmt.Sprintf("withdraw IP: %q, error: %v", vip, err))
		}
	}
}

func (controller *NetworkRoutingController) getAllVIPs(svc *corev1.Service) ([]string, []string, error) {
	return controller.getVIPs(svc, false)
}

func (controller *NetworkRoutingController) getActiveVIPs(svc *corev1.Service) ([]string, []string, error) {
	return controller.getVIPs(svc, true)
}

func (controller *NetworkRoutingController) getAllActiveVIPs() ([]string, []string, error) {
	toAdvertiseList := make([]string, 0)
	toWithdrawList := make([]string, 0)
	for _, obj := range controller.serviceLister.List() {
		svc := obj.(*corev1.Service)
		toAdvertise, toWithdraw, err := controller.getActiveVIPs(svc)
		if err != nil {
			klog.Errorf(fmt.Sprintf("svc %/%s get active vips err:%v", svc.Namespace, svc.Name, err))
			continue
		}

		toAdvertiseList = append(toAdvertiseList, toAdvertise...)
		toWithdrawList = append(toWithdrawList, toWithdraw...)
	}

	return toAdvertiseList, toWithdrawList, nil
}

func (controller *NetworkRoutingController) getVIPs(svc *corev1.Service, onlyActiveEndpoints bool) ([]string, []string, error) {
	toAdvertiseList := make([]string, 0)
	toWithdrawList := make([]string, 0)

	toAdvertise, toWithdraw, err := controller.getVIPsForService(svc, onlyActiveEndpoints)
	if err != nil {
		return nil, nil, err
	}

	if len(toAdvertise) > 0 {
		toAdvertiseList = append(toAdvertiseList, toAdvertise...)
	}

	if len(toWithdraw) > 0 {
		toWithdrawList = append(toWithdrawList, toWithdraw...)
	}

	return toAdvertiseList, toWithdrawList, nil
}

const (
	svcLocalAnnotation = "kube-router.io/service.local"
)

func (controller *NetworkRoutingController) getVIPsForService(svc *corev1.Service, onlyActiveEndpoints bool) ([]string, []string, error) {
	advertise := true
	_, hasLocalAnnotation := svc.Annotations[svcLocalAnnotation]
	hasLocalTrafficPolicy := svc.Spec.ExternalTrafficPolicy == corev1.ServiceExternalTrafficPolicyTypeLocal
	isLocal := hasLocalAnnotation || hasLocalTrafficPolicy
	if onlyActiveEndpoints && isLocal {
		var err error
		advertise, err = controller.nodeHasEndpointsForService(svc)
		if err != nil {
			return nil, nil, err
		}
	}

	ipList := controller.getAllVIPsForService(svc)

	if !advertise {
		return nil, ipList, nil
	}

	return ipList, nil, nil
}

// INFO: 如果这个 service 是 ServiceExternalTrafficPolicyTypeLocal，那 kube-router 和 endpoint 在一个 node 上， 该 service ip 才会被宣告
func (controller *NetworkRoutingController) nodeHasEndpointsForService(svc *corev1.Service) (bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(svc)
	if err != nil {
		return false, err
	}
	obj, exists, err := controller.endpointLister.GetByKey(key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("endpoint resource doesn't exist for service: %q", svc.Name)
	}
	endpoints := obj.(*corev1.Endpoints)
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			if address.NodeName != nil {
				if *address.NodeName == controller.nodeName {
					return true, nil
				}
			} else {
				if address.IP == controller.nodeIP.String() {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

const (
	svcAdvertiseClusterAnnotation      = "kube-router.io/service.advertise.clusterip"
	svcAdvertiseExternalAnnotation     = "kube-router.io/service.advertise.externalip"
	svcAdvertiseLoadBalancerAnnotation = "kube-router.io/service.advertise.loadbalancerip"
)

func (controller *NetworkRoutingController) getAllVIPsForService(svc *corev1.Service) []string {
	ipList := make([]string, 0)

	if controller.shouldAdvertiseService(svc, svcAdvertiseClusterAnnotation, controller.advertiseClusterIP) {
		if len(svc.Spec.ClusterIP) != 0 {
			ipList = append(ipList, svc.Spec.ClusterIP)
		}
	}

	if controller.shouldAdvertiseService(svc, svcAdvertiseExternalAnnotation, controller.advertiseExternalIP) {
		ipList = append(ipList, svc.Spec.ExternalIPs...)
	}

	if controller.shouldAdvertiseService(svc, svcAdvertiseLoadBalancerAnnotation, controller.advertiseLoadBalancerIP) {
		ipList = append(ipList, controller.getLoadBalancerIPs(svc)...)
	}

	return ipList
}

func (controller *NetworkRoutingController) shouldAdvertiseService(svc *corev1.Service, annotation string, defaultValue bool) bool {
	returnValue := defaultValue
	stringValue, exists := svc.Annotations[annotation]
	if exists {
		returnValue, _ = strconv.ParseBool(stringValue)
	}
	return returnValue
}

func (controller *NetworkRoutingController) getLoadBalancerIPs(svc *corev1.Service) []string {
	loadBalancerIPList := make([]string, 0)
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if len(ingress.IP) > 0 {
				loadBalancerIPList = append(loadBalancerIPList, ingress.IP)
			}
		}
	}

	return loadBalancerIPList
}
