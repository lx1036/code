package speaker

import (
	"fmt"
	"io"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/bgp"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/config"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"net"
	"strconv"
)

// Session gives access to the BGP session.
type Session interface {
	io.Closer
	Set(advs ...*bgp.Advertisement) error
}

type peer struct {
	peer *config.Peer
	bgp  Session
}

type BGPController struct {
	MyNode     string
	nodeLabels labels.Set
	peers      []*peer
	SvcAds     map[string][]*bgp.Advertisement
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

		} else if p.bgp == nil && shouldRun {
			session, err := bgp.New(net.JoinHostPort(p.peer.Addr.String(), strconv.Itoa(int(p.peer.Port))),
				p.peer.MyASN, routerID, p.peer.ASN, p.peer.HoldTime, controller.MyNode)
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

	}

	return nil
}

func (controller *BGPController) ShouldAnnounce(s string, s2 string, e *Endpoints) string {
	panic("implement me")
}

func (controller *BGPController) SetBalancer(s string, ip net.IP, pool *config.Pool) error {
	panic("implement me")
}

func (controller *BGPController) DeleteBalancer(s string, s2 string) error {
	panic("implement me")
}

func (controller *BGPController) SetNodeLabels(m map[string]string) error {
	panic("implement me")
}
