package routing

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	gobgpapi "github.com/osrg/gobgp/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"strconv"

	"k8s.io/klog/v2"
)

// INFO: ECMP(Equal Cost Multi-Path) 等价路由: 多条不同链路到达同一目的地址的网络环境，即同一个 dst 多个 next hop

func (controller *NetworkRoutingController) advertiseVIPs(vips []string) {
	for _, vip := range vips {
		klog.Infof(fmt.Sprintf("advertising route: '%s/32 via %s' to peers", vip, controller.nodeIP.String()))

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
		_, err := controller.bgpServer.AddPath(context.Background(), &gobgpapi.AddPathRequest{
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

func (controller *NetworkRoutingController) getAllVIPs() ([]string, []string, error) {
	return controller.getVIPs(false)
}

func (controller *NetworkRoutingController) getActiveVIPs() ([]string, []string, error) {
	return controller.getVIPs(true)
}

func (controller *NetworkRoutingController) getVIPs(onlyActiveEndpoints bool) ([]string, []string, error) {
	toAdvertiseList := make([]string, 0)
	toWithdrawList := make([]string, 0)

	services, err := controller.serviceLister.List(labels.Everything())
	if err != nil {
		return nil, nil, err
	}
	for _, svc := range services {
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
	endpoint, err := controller.endpointLister.Endpoints(svc.Namespace).Get(svc.Name)
	if err != nil {
		return false, err
	}

	for _, subset := range endpoint.Subsets {
		for _, address := range subset.Addresses {
			if (address.NodeName != nil && *address.NodeName == controller.nodeName) || address.IP == controller.nodeIP.String() {
				return true, nil
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
		clusterIP := controller.getClusterIP(svc)
		if len(clusterIP) != 0 {
			ipList = append(ipList, clusterIP)
		}
	}

	if controller.shouldAdvertiseService(svc, svcAdvertiseExternalAnnotation, controller.advertiseExternalIP) {
		ipList = append(ipList, controller.getExternalIPs(svc)...)
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
		// Service annotations overrides defaults.
		returnValue, _ = strconv.ParseBool(stringValue)
	}
	return returnValue
}

func (controller *NetworkRoutingController) getClusterIP(svc *corev1.Service) string {
	clusterIP := ""
	if svc.Spec.Type == corev1.ServiceTypeClusterIP ||
		svc.Spec.Type == corev1.ServiceTypeNodePort ||
		svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if !ClusterIPIsNoneOrBlank(svc.Spec.ClusterIP) {
			clusterIP = svc.Spec.ClusterIP
		}
	}

	return clusterIP
}

func (controller *NetworkRoutingController) getExternalIPs(svc *corev1.Service) []string {
	externalIPList := make([]string, 0)
	if svc.Spec.Type == corev1.ServiceTypeClusterIP ||
		svc.Spec.Type == corev1.ServiceTypeNodePort ||
		svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		// skip headless services
		if !ClusterIPIsNoneOrBlank(svc.Spec.ClusterIP) {
			externalIPList = append(externalIPList, svc.Spec.ExternalIPs...)
		}
	}

	return externalIPList
}

func (controller *NetworkRoutingController) getLoadBalancerIPs(svc *corev1.Service) []string {
	loadBalancerIPList := make([]string, 0)
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		// skip headless services
		if !ClusterIPIsNoneOrBlank(svc.Spec.ClusterIP) {
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				if len(ingress.IP) > 0 {
					loadBalancerIPList = append(loadBalancerIPList, ingress.IP)
				}
			}
		}
	}

	return loadBalancerIPList
}

func ClusterIPIsNoneOrBlank(clusterIP string) bool {
	return clusterIP == corev1.ClusterIPNone || len(clusterIP) == 0
}
