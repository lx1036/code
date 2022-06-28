package app

import (
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

	if err := d.serviceBPFManager.InitMaps(option.Config.EnableIPv6, option.Config.EnableIPv4,
		createSockRevNatMaps, option.Config.RestoreState); err != nil {
		log.WithError(err).Fatal("Unable to initialize service maps")
	}

}
