package proxy

import (
	"fmt"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"net"

	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps/cachingmap"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps/nat"
)

type Option func(Proxy) error

type KubeProxy struct {
	hostIPUpdates chan []net.IP
}

func StartKubeProxy(k8s kubernetes.Interface, hostname string,
	bpfMapContext *bpf.MapContext, opts ...Option) {

	p := &KubeProxy{
		hostIPUpdates: make(chan []net.IP, 1),
	}

	go func() {
		err := p.start()
		if err != nil {
			log.WithError(err).Panic("kube-proxy failed to start")
		}
	}()

}

func (p *KubeProxy) start() error {
	// wait for the initial update
	hostIPs := <-p.hostIPUpdates
	err := p.run(hostIPs)
	if err != nil {
		return err
	}

	go func() {
		for {
			hostIPs, ok := <-p.hostIPUpdates
			if !ok {
				p.proxy.Stop()
				return
			}

		}
	}()

}

func (p *KubeProxy) run(hostIPs []net.IP) error {

	feCache := cachingmap.New(nat.FrontendMapParameters, p.frontendMap)
	beCache := cachingmap.New(nat.BackendMapParameters, p.backendMap)
	syncer, err := NewSyncer(withLocalNP, feCache, beCache, p.affinityMap, p.rt)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("new bpf syncer err: %v", err))
	}

	proxy, err := New(p.k8s, syncer, p.hostname, p.opts...)
	if err != nil {
		return err
	}

	log.Infof("kube-proxy started, hostname=%q hostIPs=%+v", p.hostname, hostIPs)

	p.proxy = proxy
	p.syncer = syncer

	return nil
}

// OnHostIPsUpdate 该函数会被外部触发调用
func (p *KubeProxy) OnHostIPsUpdate(ips []net.IP) {
	select {
	case p.hostIPUpdates <- ips:
	default:
		// if block, drop the stale and replace with new one
		select {
		case <-p.hostIPUpdates:
		default:
		}
		p.hostIPUpdates <- ips
	}
	klog.Infof(fmt.Sprintf("kube-proxy OnHostIPsUpdate %+v", ips))
}
