package iptables

import "github.com/cilium/cilium/pkg/lock"

type IptablesManager struct {
	// This lock ensures there are no concurrent executions of the InstallRules() and
	// InstallProxyRules() methods, as otherwise we may end up with errors (as rules may have
	// been already removed or installed by a different execution of the method) or with an
	// inconsistent ruleset
	lock.Mutex

	haveSocketMatch      bool
	haveBPFSocketAssign  bool
	ipEarlyDemuxDisabled bool
}
