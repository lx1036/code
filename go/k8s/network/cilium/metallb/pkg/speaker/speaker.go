package speaker

import (
	"fmt"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/k8s"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/k8s/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"net"

	"k8s-lx1036/k8s/network/cilium/metallb/pkg/bgp"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/config"
)

// A Protocol can advertise an IP address.
type Protocol interface {
	SetConfig(*config.Config) error
	ShouldAnnounce(string, string, *Endpoints) string
	SetBalancer(string, net.IP, *config.Pool) error
	DeleteBalancer(string, string) error
	SetNodeLabels(map[string]string) error
}

type Config struct {
	MyNode string
	SList  SpeakerList
}

type Speaker struct {
	myNode string

	config *config.Config
	Client service

	protocols map[config.Proto]Protocol
	announced map[string]config.Proto // service name -> protocol advertising it
	svcIP     map[string]net.IP       // service name -> assigned IP
}

func NewSpeaker(cfg Config) (*Speaker, error) {
	protocols := map[config.Proto]Protocol{
		config.BGP: &BGPController{
			MyNode: cfg.MyNode,
			SvcAds: make(map[string][]*bgp.Advertisement),
		},
	}

	ret := &Speaker{
		myNode:    cfg.MyNode,
		protocols: protocols,
		announced: map[string]config.Proto{},
		svcIP:     map[string]net.IP{},
	}

	return ret, nil
}

func (speaker *Speaker) SetConfig(cfg *config.Config) types.SyncState {
	if cfg == nil {
		klog.Errorf(fmt.Sprintf("[SetConfig]config is required"))
		return types.SyncStateError
	}

	// 检查 svcIP 是否在新的 pool 配置中
	for svc, ip := range speaker.svcIP {
		if pool := poolFor(cfg.Pools, ip); pool == "" {
			klog.Errorf(fmt.Sprintf("service %s has no configuration under new config", svc))
			return types.SyncStateError
		}
	}

	for proto, handler := range speaker.protocols {
		if err := handler.SetConfig(cfg); err != nil {
			klog.Errorf(fmt.Sprintf("[SetConfig]applying new configuration to protocol %s handler failed", proto))
			return types.SyncStateError
		}
	}

	speaker.config = cfg

	return types.SyncStateReprocessAll
}

// Service represents an object containing the minimal representation of a
// v1.Service object needed for announcements.
type Service struct {
	Type          string
	TrafficPolicy string
	Ingress       []corev1.LoadBalancerIngress
}

func (speaker *Speaker) SetBalancer(name string, svc *corev1.Service, eps *corev1.Endpoints) types.SyncState {
	s := speaker.SetService(name, &Service{
		Type:          string(svc.Spec.Type),
		TrafficPolicy: string(svc.Spec.ExternalTrafficPolicy),
		Ingress:       svc.Status.LoadBalancer.Ingress,
	}, toEndpoints(eps))
	if s == types.SyncStateSuccess {
		klog.Infof(fmt.Sprintf("announcing from node %q", speaker.myNode))
	}
	return s
}

func (speaker *Speaker) SetService(name string, svc *Service, eps *Endpoints) types.SyncState {
	if svc == nil {
		return speaker.deleteBalancer(name, "serviceDeleted")
	}

	if svc.Type != string(corev1.ServiceTypeLoadBalancer) {
		return speaker.deleteBalancer(name, "notLoadBalancer")
	}

	if speaker.config == nil {
		return types.SyncStateSuccess
	}

	if len(svc.Ingress) != 1 {
		return speaker.deleteBalancer(name, "noIPAllocated")
	}

	lbIP := net.ParseIP(svc.Ingress[0].IP)
	if lbIP == nil {
		return speaker.deleteBalancer(name, "invalidIP")
	}
	poolName := poolFor(speaker.config.Pools, lbIP)
	if poolName == "" {
		return speaker.deleteBalancer(name, "ipNotAllowed")
	}

}

func poolFor(pools map[string]*config.Pool, ip net.IP) string {
	for pname, p := range pools {
		for _, cidr := range p.CIDR {
			if cidr.Contains(ip) {
				return pname
			}
		}
	}
	return ""
}
