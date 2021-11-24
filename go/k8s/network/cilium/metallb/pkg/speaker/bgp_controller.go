package speaker

import (
	"io"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/bgp"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/config"
	"k8s.io/apimachinery/pkg/labels"
	"net"
)

// Session gives access to the BGP session.
type Session interface {
	io.Closer
	Set(advs ...*bgp.Advertisement) error
}

type peer struct {
	cfg *config.Peer
	bgp Session
}

type BGPController struct {
	MyNode     string
	nodeLabels labels.Set
	peers      []*peer
	SvcAds     map[string][]*bgp.Advertisement
}

// SetConfig INFO: 会立即和 router server 建立 bgp session
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
