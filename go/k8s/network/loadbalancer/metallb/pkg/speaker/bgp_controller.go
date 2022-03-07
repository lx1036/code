package speaker

import (
	"fmt"
	"k8s-lx1036/k8s/network/loadbalancer/metallb/pkg/bgp"
	"k8s-lx1036/k8s/network/loadbalancer/metallb/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"net"
	"sort"
	"strconv"
)

type peer struct {
	peer *config.Peer
	bgp  *bgp.Session
}

type BGPController struct {
	MyNode     string
	nodeLabels labels.Set
	peers      []*peer
	svcAds     map[string][]*bgp.Advertisement
}

// SetConfig INFO: 会立即和 router server 建立 bgp session
func (controller *BGPController) SetConfig(cfg *config.Config) error {
	newPeers := make([]*peer, 0, len(cfg.Peers))
	for _, p := range cfg.Peers {
		newPeers = append(newPeers, &peer{
			peer: p,
		})
	}

	controller.peers = newPeers

	return controller.syncPeers()
}

// INFO: new peer 会立即与 router server 建立 bgp session 连接
func (controller *BGPController) syncPeers() error {
	var needUpdateAds bool
	for _, p := range controller.peers {
		// First, determine if the peering should be active for this node
		shouldRun := false
		for _, ns := range p.peer.NodeSelectors {
			if ns.Matches(controller.nodeLabels) {
				shouldRun = true
				break
			}
		}

		if p.bgp != nil && !shouldRun {
			klog.Infof("peer deconfigured, closing BGP session")
			if err := p.bgp.Close(); err != nil {
				klog.Infof(fmt.Sprintf("[syncPeers]failed to shut down BGP session:%v", err))
			}
			p.bgp = nil
		} else if p.bgp == nil && shouldRun {
			klog.Infof("peer configured, starting BGP session")
			var routerID net.IP
			if p.peer.RouterID != nil {
				routerID = p.peer.RouterID
			}
			session, err := bgp.New(net.JoinHostPort(p.peer.Addr.String(), strconv.Itoa(int(p.peer.Port))),
				routerID, p.peer.ASN, p.peer.MyASN, p.peer.HoldTime, controller.MyNode)
			if err != nil {
				klog.Errorf(fmt.Sprintf("[BGPController syncPeers]failed to establish BGP session with router server %s/%d",
					p.peer.Addr.String(), p.peer.Port))
				return err
			} else {
				p.bgp = session
				needUpdateAds = true
			}
		}
	}

	if needUpdateAds {
		if err := controller.updateLoadBalancerAds(); err != nil {
			klog.Errorf("failed to update BGP advertisements")
			return err
		}
	}

	return nil
}

// INFO: 宣告 LoadBalancer Service IP to router server
func (controller *BGPController) updateLoadBalancerAds() error {
	var allAds []*bgp.Advertisement
	for _, ads := range controller.svcAds {
		allAds = append(allAds, ads...)
	}

	for _, session := range controller.PeerSessions() {
		if session == nil {
			continue
		}
		if err := session.AddPath(allAds...); err != nil {
			return err
		}
	}

	return nil
}

func (controller *BGPController) PeerSessions() []*bgp.Session {
	s := make([]*bgp.Session, len(controller.peers))
	for i, peer := range controller.peers {
		s[i] = peer.bgp
	}

	return s
}

// ShouldAnnounce
// INFO: externalTrafficPolicy=Cluster && any healthy endpoint exists || externalTrafficPolicy=Local && there's a ready local endpoint
//  如果是 Cluster，只需要检查有 healthy endpoint，不管在不在 speaker 生效的 node 上；如果是 Local，则必须 speaker 生效的 node 上有 healthy endpoint
func (controller *BGPController) ShouldAnnounce(name string, policyType string, eps *Endpoints) (string, bool) {
	switch corev1.ServiceExternalTrafficPolicyType(policyType) {
	case corev1.ServiceExternalTrafficPolicyTypeLocal:
		for _, endpoint := range eps.Ready {
			if *endpoint.NodeName == controller.MyNode {
				return "", true
			}
		}
		return "noLocalHealthyEndpoints", false
	case corev1.ServiceExternalTrafficPolicyTypeCluster:
		if len(eps.Ready) != 0 {
			return "", true
		}
		return "noHealthyEndpoints", false
	default:
		return "unknownTrafficPolicy", false
	}
}

func (controller *BGPController) SetBalancer(name string, lbIP net.IP, pool *config.Pool) error {
	controller.svcAds[name] = nil
	for _, adCfg := range pool.BGPAdvertisements {
		m := net.CIDRMask(adCfg.AggregationLength, 32)
		ad := &bgp.Advertisement{
			Prefix: &net.IPNet{
				IP:   lbIP.Mask(m),
				Mask: m,
			},
			LocalPref: adCfg.LocalPref,
			NextHop:   net.ParseIP("10.20.30.40"),
			//LocalPref:   42,
			Communities: []uint32{1234, 2345},
		}
		for comm := range adCfg.Communities {
			ad.Communities = append(ad.Communities, comm)
		}
		sort.Slice(ad.Communities, func(i, j int) bool {
			return ad.Communities[i] < ad.Communities[j]
		})
		controller.svcAds[name] = append(controller.svcAds[name], ad)
	}

	if err := controller.updateLoadBalancerAds(); err != nil {
		klog.Errorf("failed to update BGP advertisements")
		return err
	}

	return nil
}

func (controller *BGPController) DeleteBalancer(s string, s2 string) error {
	panic("implement me")
}

func (controller *BGPController) SetNodeLabels(m map[string]string) error {
	panic("implement me")
}
