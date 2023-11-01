package service

import (
	"net"

	"github.com/cilium/cilium/pkg/service/healthserver"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/loadbalancer"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/testutils/mockmaps"

	. "gopkg.in/check.v1"
)

var (
	frontend1 = *loadbalancer.NewL3n4AddrID(loadbalancer.TCP, net.ParseIP("1.1.1.1"), 80, loadbalancer.ScopeExternal, 0)
	frontend2 = *loadbalancer.NewL3n4AddrID(loadbalancer.TCP, net.ParseIP("1.1.1.2"), 80, loadbalancer.ScopeExternal, 0)
	frontend3 = *loadbalancer.NewL3n4AddrID(loadbalancer.TCP, net.ParseIP("f00d::1"), 80, loadbalancer.ScopeExternal, 0)
	backends1 = []loadbalancer.Backend{
		*loadbalancer.NewBackend(0, loadbalancer.TCP, net.ParseIP("10.0.0.1"), 8080),
		*loadbalancer.NewBackend(0, loadbalancer.TCP, net.ParseIP("10.0.0.2"), 8080),
	}
	backends2 = []loadbalancer.Backend{
		*loadbalancer.NewBackend(0, loadbalancer.TCP, net.ParseIP("10.0.0.2"), 8080),
		*loadbalancer.NewBackend(0, loadbalancer.TCP, net.ParseIP("10.0.0.3"), 8080),
	}
	backends3 = []loadbalancer.Backend{
		*loadbalancer.NewBackend(0, loadbalancer.TCP, net.ParseIP("fd00::2"), 8080),
		*loadbalancer.NewBackend(0, loadbalancer.TCP, net.ParseIP("fd00::3"), 8080),
	}
)

type ManagerTestSuite struct {
	svc                       *Service
	lbmap                     *mockmaps.LBMockMap // for accessing public fields
	svcHealth                 *healthserver.MockHealthHTTPServerFactory
	prevOptionSessionAffinity bool
	prevOptionLBSourceRanges  bool
	prevOptionNPAlgo          string
	prevOptionDPMode          string
	ipv6                      bool
}

var _ = Suite(&ManagerTestSuite{})

func (m *ManagerTestSuite) SetUpTest(c *C) {

}

func (m *ManagerTestSuite) TearDownTest(c *C) {

}
