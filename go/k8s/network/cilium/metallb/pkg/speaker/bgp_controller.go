package speaker

import (
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/bgp"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/config"
	"k8s.io/apimachinery/pkg/labels"
	"net"
)

type BGPController struct {
	MyNode     string
	nodeLabels labels.Set
	peers      []*peer
	SvcAds     map[string][]*bgp.Advertisement
}

func (bgp *BGPController) SetConfig(config *config.Config) error {
	panic("implement me")
}

func (bgp *BGPController) ShouldAnnounce(s string, s2 string, e *interface{}) string {
	panic("implement me")
}

func (bgp *BGPController) SetBalancer(s string, ip net.IP, pool *config.Pool) error {
	panic("implement me")
}

func (bgp *BGPController) DeleteBalancer(s string, s2 string) error {
	panic("implement me")
}

func (bgp *BGPController) SetNodeLabels(m map[string]string) error {
	panic("implement me")
}
