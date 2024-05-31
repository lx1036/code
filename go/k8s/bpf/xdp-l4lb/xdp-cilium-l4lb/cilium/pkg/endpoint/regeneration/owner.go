package regeneration

import (
    "context"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/fqdn/restore"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
    monitorAPI "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/api"
)

// Owner is the interface defines the requirements for anybody owning policies.
// cmd/Daemon 对象实现该接口
type Owner interface {
    // QueueEndpointBuild puts the given endpoint in the processing queue
    QueueEndpointBuild(ctx context.Context, epID uint64) (func(), error)

    // GetCompilationLock returns the mutex responsible for synchronizing compilation
    // of BPF programs.
    GetCompilationLock() *lock.RWMutex

    // GetCIDRPrefixLengths returns the sorted list of unique prefix lengths used
    // by CIDR policies.
    GetCIDRPrefixLengths() (s6, s4 []int)

    // SendNotification is called to emit an agent notification
    SendNotification(msg monitorAPI.AgentNotifyMessage) error

    // Datapath returns a reference to the datapath implementation.
    Datapath() datapath.Datapath

    // GetDNSRules creates a fresh copy of DNS rules that can be used when
    // endpoint is restored on a restart.
    GetDNSRules(epID uint16) restore.DNSRules

    // RemoveRestoredDNSRules removes any restored DNS rules for
    // this endpoint from the DNS proxy.
    RemoveRestoredDNSRules(epID uint16)
}
