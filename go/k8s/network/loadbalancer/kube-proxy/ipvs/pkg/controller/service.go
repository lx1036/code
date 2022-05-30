package controller

import (
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/moby/ipvs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	svcDSRAnnotation                = "kube-router.io/service.dsr"
	svcSchedulerAnnotation          = "kube-router.io/service.scheduler"
	svcHairpinAnnotation            = "kube-router.io/service.hairpin"
	svcHairpinExternalIPsAnnotation = "kube-router.io/service.hairpin.externalips"
	svcLocalAnnotation              = "kube-router.io/service.local"

	// Taken from https://github.com/torvalds/linux/blob/master/include/uapi/linux/ip_vs.h#L21
	ipvsPersistentFlagHex = 0x0001
	ipvsHashedFlagHex     = 0x0002
	ipvsOnePacketFlagHex  = 0x0004
	ipvsSched1FlagHex     = 0x0008
	ipvsSched2FlagHex     = 0x0010
	ipvsSched3FlagHex     = 0x0020
)

func (controller *NetworkServiceController) onServiceUpdate(service *corev1.Service) {
	controller.Lock()
	defer controller.Unlock()

	if !controller.readyForUpdates {
		return
	}

	newSvcMap := controller.buildSvcInfo()
	newEndpointMap := controller.buildEndpointInfo()
	if !newSvcMap.equal(controller.serviceMap) {
		controller.serviceMap = newSvcMap
		controller.endpointMap = newEndpointMap
		controller.sync()
	} else {
		klog.Infof(fmt.Sprintf("skipping syncing IPVS services for update to service %s/%s as nothing changed", service.Namespace, service.Name))
	}
}

// INFO: 从 Service annotation 抽取出配置信息
type serviceInfo struct {
	name      string
	namespace string

	// ClusterIP/NodePort
	address    net.IP
	port       int
	targetPort string
	protocol   string
	nodePort   int

	flags                         uint32
	sessionAffinity               bool
	sessionAffinityTimeoutSeconds uint32
	isLocal                       bool

	directServerReturn       bool
	directServerReturnMethod string

	scheduler string

	hairpin            bool
	hairpinExternalIPs bool
	skipLbIps          bool
	externalIPs        []string
	loadBalancerIPs    []string
}

type serviceInfoMap map[string]*serviceInfo

func (svc serviceInfoMap) equal(other serviceInfoMap) bool {
	if len(svc) != len(other) {
		return false
	}

	return reflect.DeepEqual(svc, other)
}

func (controller *NetworkServiceController) buildSvcInfo() serviceInfoMap {
	svcInfoMap := make(serviceInfoMap)

	for _, obj := range controller.svcLister.List() {
		svc := obj.(*corev1.Service)
		for _, port := range svc.Spec.Ports {
			svcInfo := serviceInfo{
				name:            svc.ObjectMeta.Name,
				namespace:       svc.ObjectMeta.Namespace,
				address:         net.ParseIP(svc.Spec.ClusterIP),
				port:            int(port.Port),
				targetPort:      port.TargetPort.String(),
				protocol:        strings.ToLower(string(port.Protocol)),
				nodePort:        int(port.NodePort),
				externalIPs:     svc.Spec.ExternalIPs,
				isLocal:         false,
				sessionAffinity: svc.Spec.SessionAffinity == corev1.ServiceAffinityClientIP,
			}
			if svcInfo.sessionAffinity { // INFO: service 是 client session affinity 的，则不是轮询选择下一跳endpoint，而是一直都是同一个 endpoint
				// Kube-apiserver side guarantees SessionAffinityConfig won't be nil when session affinity
				// type is ClientIP
				// https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/core/v1/defaults.go#L106
				svcInfo.sessionAffinityTimeoutSeconds = uint32(*svc.Spec.SessionAffinityConfig.ClientIP.TimeoutSeconds)
				svcInfo.flags |= ipvsPersistentFlagHex
			} else {
				svcInfo.sessionAffinityTimeoutSeconds = 0
				svcInfo.flags &^= ipvsPersistentFlagHex
			}
			for _, lbIngress := range svc.Status.LoadBalancer.Ingress {
				if len(lbIngress.IP) > 0 {
					svcInfo.loadBalancerIPs = append(svcInfo.loadBalancerIPs, lbIngress.IP)
				}
			}

			_, svcInfo.hairpin = svc.ObjectMeta.Annotations[svcHairpinAnnotation]
			_, svcInfo.hairpinExternalIPs = svc.ObjectMeta.Annotations[svcHairpinExternalIPsAnnotation]
			_, svcInfo.isLocal = svc.ObjectMeta.Annotations[svcLocalAnnotation]
			if svc.Spec.ExternalTrafficPolicy == corev1.ServiceExternalTrafficPolicyTypeLocal {
				svcInfo.isLocal = true // INFO: local externalTrafficPolicy 表示该 svc->eps, 只考虑本机上的 ep 才会作为 ipvs rs
			}

			// ipvs schedule 负载均衡算法
			dsrMethod, ok := svc.ObjectMeta.Annotations[svcDSRAnnotation]
			if ok {
				svcInfo.directServerReturn = true
				svcInfo.directServerReturnMethod = dsrMethod
			}
			svcInfo.scheduler = ipvs.RoundRobin
			schedulingMethod, ok := svc.ObjectMeta.Annotations[svcSchedulerAnnotation]
			if ok {
				switch {
				case schedulingMethod == ipvs.RoundRobin:
					svcInfo.scheduler = ipvs.RoundRobin
				case schedulingMethod == ipvs.LeastConnection:
					svcInfo.scheduler = ipvs.LeastConnection
				case schedulingMethod == ipvs.DestinationHashing:
					svcInfo.scheduler = ipvs.DestinationHashing
				case schedulingMethod == ipvs.SourceHashing:
					svcInfo.scheduler = ipvs.SourceHashing
				}
			}

			svcID := generateServiceID(svc.Namespace, svc.Name, port.Name)
			svcInfoMap[svcID] = &svcInfo
		}
	}

	return svcInfoMap
}

// 多个 port 时需要有 name，且 name 不可以重复
func generateServiceID(namespace, svcName, port string) string {
	return namespace + "-" + svcName + "-" + port
}

func IsHeadlessService(svc *corev1.Service) bool {
	return svc.Spec.Type == corev1.ServiceTypeClusterIP &&
		(svc.Spec.ClusterIP == corev1.ClusterIPNone || len(svc.Spec.ClusterIP) == 0)
}

func IsExternalNameService(svc *corev1.Service) bool {
	return svc.Spec.Type == corev1.ServiceTypeExternalName
}
