package loader

import (
	"context"
	"fmt"
	"github.com/cilium/cilium/pkg/common"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/node"
	"github.com/cilium/cilium/pkg/sysctl"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/cgroup"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/defaults"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/option"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath"

	"github.com/vishvananda/netlink"
)

const (
	initArgLib int = iota
	initArgRundir
	initArgIPv4NodeIP
	initArgIPv6NodeIP
	initArgMode
	initArgDevices
	initArgXDPDevice
	initArgXDPMode
	initArgMTU
	initArgIPSec
	initArgEncryptInterface
	initArgHostReachableServices
	initArgHostReachableServicesUDP
	initArgHostReachableServicesPeer
	initArgCgroupRoot
	initArgBpffsRoot
	initArgNodePort
	initArgNodePortBind
	initBPFCPU
	initArgNodePortIPv4Addrs
	initArgNodePortIPv6Addrs
	initArgNrCPUs
	initArgMax
)

// firstInitialization is true when Reinitialize() is called for the first
// time. It can only be accessed when GetCompilationLock() is being held.
var firstInitialization = true

// Loader is a wrapper structure around operations related to compiling,
// loading, and reloading datapath programs.
type Loader struct {
}

// NewLoader returns a new loader.
func NewLoader() *Loader {
	return &Loader{}
}

// Reinitialize (re-)configures the base datapath configuration including global
// BPF programs, netfilter rule configuration and reserving routes in IPAM for
// locally detected prefixes. It may be run upon initial Cilium startup, after
// restore from a previous Cilium run, or during regular Cilium operation.
func (l *Loader) Reinitialize(ctx context.Context, o datapath.BaseProgramOwner, deviceMTU int, iptMgr datapath.IptablesManager, p datapath.Proxy, r datapath.RouteReserver) error {
	sysSettings := []setting{
		{"net.core.bpf_jit_enable", "1", true},
		{"net.ipv4.conf.all.rp_filter", "0", false},
		{"kernel.unprivileged_bpf_disabled", "1", true},
	}
	for _, s := range sysSettings {
		log.Infof("Setting sysctl %s=%s", s.name, s.val)
		if err := sysctl.Write(s.name, s.val); err != nil {
			if !s.ignoreErr {
				return fmt.Errorf("Failed to sysctl -w %s=%s: %s", s.name, s.val, err)
			}
			log.WithError(err).WithFields(log.Fields{
				logfields.SysParamName:  s.name,
				logfields.SysParamValue: s.val,
			}).Warning("Failed to sysctl -w")
		}
	}

	// INFO: 执行 bpf init.sh
	args := make([]string, initArgMax)
	args[initArgLib] = option.Config.BpfDir
	args[initArgRundir] = option.Config.StateDir
	args[initArgCgroupRoot] = cgroup.GetCgroupRoot()
	args[initArgBpffsRoot] = bpf.GetMapRoot()
	args[initArgMTU] = fmt.Sprintf("%d", deviceMTU)
	args[initBPFCPU] = GetBPFCPU()
	args[initArgNrCPUs] = fmt.Sprintf("%d", common.GetNumPossibleCPUs(log))
	if option.Config.EnableIPv4 {
		args[initArgIPv4NodeIP] = node.GetInternalIPv4().String()
	} else {
		args[initArgIPv4NodeIP] = "<nil>"
	}
	if option.Config.EnableIPSec {
		args[initArgIPSec] = "true"
	} else {
		args[initArgIPSec] = "false"
	}
	if option.Config.EnableHostReachableServices {
		args[initArgHostReachableServices] = "true"
		if option.Config.EnableHostServicesUDP {
			args[initArgHostReachableServicesUDP] = "true"
		} else {
			args[initArgHostReachableServicesUDP] = "false"
		}
		if option.Config.EnableHostServicesPeer {
			args[initArgHostReachableServicesPeer] = "true"
		} else {
			args[initArgHostReachableServicesPeer] = "false"
		}
	} else {
		args[initArgHostReachableServices] = "false"
		args[initArgHostReachableServicesUDP] = "false"
		args[initArgHostReachableServicesPeer] = "false"
	}
	if len(option.Config.Devices) != 0 {
		for _, device := range option.Config.Devices {
			_, err := netlink.LinkByName(device)
			if err != nil {
				log.WithError(err).WithField("device", device).Warn("Link does not exist")
				return err
			}
		}
		if option.Config.Tunnel != option.TunnelDisabled {
			args[initArgMode] = option.Config.Tunnel
		} else if option.Config.DatapathMode == datapathOption.DatapathModeIpvlan {
			args[initArgMode] = "ipvlan"
		} else {
			args[initArgMode] = "direct"
		}
		args[initArgDevices] = strings.Join(option.Config.Devices, ";")
	} else {
		args[initArgMode] = option.Config.Tunnel
		args[initArgDevices] = "<nil>"
		if option.Config.IsFlannelMasterDeviceSet() {
			args[initArgMode] = "flannel"
			args[initArgDevices] = option.Config.FlannelMasterDevice
		}
	}
	if option.Config.EnableEndpointRoutes == true {
		args[initArgMode] = "routed"
	}
	if option.Config.EnableNodePort {
		args[initArgNodePort] = "true"
		if option.Config.EnableIPv4 {
			addrs := node.GetNodePortIPv4AddrsWithDevices()
			tmp := make([]string, 0, len(addrs))
			for iface, ipv4 := range addrs {
				tmp = append(tmp,
					fmt.Sprintf("%s=%#x", iface,
						byteorder.HostSliceToNetwork(ipv4, reflect.Uint32).(uint32)))
			}
			args[initArgNodePortIPv4Addrs] = strings.Join(tmp, ";")
		} else {
			args[initArgNodePortIPv4Addrs] = "<nil>"
		}
		args[initArgNodePortIPv6Addrs] = "<nil>"
	} else {
		args[initArgNodePort] = "false"
		args[initArgNodePortIPv4Addrs] = "<nil>"
		args[initArgNodePortIPv6Addrs] = "<nil>"
	}
	if option.Config.NodePortBindProtection {
		args[initArgNodePortBind] = "true"
	} else {
		args[initArgNodePortBind] = "false"
	}
	// INFO: @see bpf/init.sh shell 文件, 执行 ./init.sh
	prog := filepath.Join(option.Config.BpfDir, "init.sh")
	ctx, cancel := context.WithTimeout(ctx, defaults.ExecTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Env = bpf.Environment()
	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}

	if err := o.Datapath().Node().NodeConfigurationChanged(*o.LocalConfig()); err != nil {
		return err
	}

	// INFO: install iptables rules
	if option.Config.InstallIptRules {
		if err := iptMgr.TransientRulesStart(option.Config.HostDevice); err != nil {
			log.WithError(err).Warning("failed to install transient iptables rules")
		}
	}
	// The iptables rules are only removed on the first initialization to
	// remove stale rules or when iptables is enabled. The first invocation
	// is silent as rules may not exist.
	if firstInitialization || option.Config.InstallIptRules {
		iptMgr.RemoveRules(firstInitialization)
	}
	if option.Config.InstallIptRules {
		err := iptMgr.InstallRules(option.Config.HostDevice)
		iptMgr.TransientRulesEnd(false)
		if err != nil {
			return err
		}
	}

	// Reinstall proxy rules for any running proxies
	if p != nil {
		p.ReinstallRules()
	}

	return nil
}
