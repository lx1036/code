package app

import (
	"fmt"
	"github.com/cilium/cilium/pkg/ipam"
	log "github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
)

func (d *Daemon) allocateIPs() error {

	routerIP, err := d.allocateDatapathIPs(d.datapath.LocalNodeAddressing().IPv4())
	if err != nil {
		return err
	}
	if routerIP != nil {
		node.SetInternalIPv4(routerIP)
	}

	return d.allocateHealthIPs()
}

func (d *Daemon) allocateHealthIPs() error {
	result, err := d.ipam.AllocateNextFamilyWithoutSyncUpstream(ipam.IPv4, "health")
	if err != nil {
		return fmt.Errorf("unable to allocate health IPs: %s,see https://cilium.link/ipam-range-full", err)
	}

	klog.Infof("IPv4 health endpoint address: %s", result.IP)
	d.nodeDiscovery.LocalNode.IPv4HealthIP = result.IP

	// In ENI mode, we require the gateway, CIDRs, and the ENI MAC addr
	// in order to set up rules and routes on the local node to direct
	// endpoint traffic out of the ENIs.
	if option.Config.IPAM == ipamOption.IPAMENI {
		if err := d.parseHealthEndpointInfo(result); err != nil {
			log.WithError(err).Warn("Unable to allocate health information for ENI")
		}
	}

	return nil
}
