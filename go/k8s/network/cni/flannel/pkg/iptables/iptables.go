package iptables

import (
	"fmt"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/ip"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/subnet"
	"k8s.io/klog/v2"

	coreosIptables "github.com/coreos/go-iptables/iptables"
)

type IPTablesRule struct {
	table    string
	chain    string
	rulespec []string
}

func MasqRules(ipn ip.IP4Net, lease *subnet.Lease) []IPTablesRule {
	supportsRandomFully := false
	ipt, err := coreosIptables.New()
	if err == nil {
		supportsRandomFully = ipt.HasRandomFully()
	}

	n := ipn.String()           // "Network": "10.244.0.0/16"
	sn := lease.Subnet.String() // TODO: ???
	iptablesRules := []IPTablesRule{
		// This rule makes sure we don't NAT traffic within overlay network (e.g. coming out of docker0)
		{"nat", "POSTROUTING", []string{"-s", n, "-d", n, "-m", "comment", "--comment", "flanneld masq", "-j", "RETURN"}},
		// Prevent performing Masquerade on external traffic which arrives from a Node that owns the container/pod IP address
		{"nat", "POSTROUTING", []string{"!", "-s", n, "-d", sn, "-m", "comment", "--comment", "flanneld masq", "-j", "RETURN"}},
	}
	if supportsRandomFully {
		iptablesRules = append(iptablesRules,
			// NAT if it's not multicast traffic
			IPTablesRule{"nat", "POSTROUTING", []string{"-s", n, "!", "-d", "224.0.0.0/4", "-m", "comment", "--comment", "flanneld masq", "-j", "MASQUERADE", "--random-fully"}},
			// Masquerade anything headed towards flannel from the host
			IPTablesRule{"nat", "POSTROUTING", []string{"!", "-s", n, "-d", n, "-m", "comment", "--comment", "flanneld masq", "-j", "MASQUERADE", "--random-fully"}})
	} else {
		iptablesRules = append(iptablesRules,
			// NAT if it's not multicast traffic
			IPTablesRule{"nat", "POSTROUTING", []string{"-s", n, "!", "-d", "224.0.0.0/4", "-m", "comment", "--comment", "flanneld masq", "-j", "MASQUERADE"}},
			// Masquerade anything headed towards flannel from the host
			IPTablesRule{"nat", "POSTROUTING", []string{"!", "-s", n, "-d", n, "-m", "comment", "--comment", "flanneld masq", "-j", "MASQUERADE"}})
	}

	return iptablesRules
}

func ForwardRules(flannelNetwork string) []IPTablesRule {
	return []IPTablesRule{
		// These rules allow traffic to be forwarded if it is to or from the flannel network range.
		{"filter", "FORWARD", []string{"-s", flannelNetwork, "-m", "comment", "--comment", "flanneld forward", "-j", "ACCEPT"}},
		{"filter", "FORWARD", []string{"-d", flannelNetwork, "-m", "comment", "--comment", "flanneld forward", "-j", "ACCEPT"}},
	}
}

// SetupAndEnsureIP4Tables sync iptables rules
func SetupAndEnsureIP4Tables(rules []IPTablesRule, resyncPeriod int) {
	ipt, err := coreosIptables.New()
	if err != nil {
		// if we can't find iptables, give up and return
		klog.Errorf(fmt.Sprintf("Failed to setup IPTables. iptables binary was not found: %v", err))
		return
	}

}
