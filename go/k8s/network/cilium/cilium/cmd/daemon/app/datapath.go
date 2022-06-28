package app

import (
	"github.com/cilium/cilium/pkg/maps/ctmap"
	log "github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
	"os"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
)

// initMaps opens all BPF maps (and creates them if they do not exist). This
// must be done *before* any operations which read BPF maps, especially
// restoring endpoints and services.
func (d *Daemon) initMaps() error {

	// Delete old maps if left over from an upgrade.
	for _, name := range []string{"cilium_proxy4", "cilium_proxy6", "cilium_policy"} {
		path := bpf.MapPath(name)
		if _, err := os.Stat(path); err == nil {
			if err = os.RemoveAll(path); err == nil {
				klog.Infof("removed legacy map file %s", path)
			}
		}
	}

	// INFO: init service bpf maps
	createSockRevNatMaps := true
	if err := d.serviceBPFManager.InitMaps(false, true, createSockRevNatMaps, true); err != nil {
		log.WithError(err).Fatal("Unable to initialize service maps")
	}

	// INFO: init endpoint bpf maps
	for _, ep := range d.endpointManager.GetEndpoints() {
		ep.InitMap()
	}
	for _, ep := range d.endpointManager.GetEndpoints() {
		if !ep.ConntrackLocal() {
			continue
		}
		for _, m := range ctmap.LocalMaps(ep, option.Config.EnableIPv4,
			option.Config.EnableIPv6) {
			if _, err := m.Create(); err != nil {
				return err
			}
		}
	}
	for _, m := range ctmap.GlobalMaps(option.Config.EnableIPv4,
		option.Config.EnableIPv6) {
		if _, err := m.Create(); err != nil {
			return err
		}
	}

}
