package controller

import (
	"bytes"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"k8s.io/client-go/tools/cache"
	"net"
	"strings"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/network-policy/pkg/ipset"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/util/iptables"
	"k8s.io/utils/exec"
)

const (
	kubeInputChainName     = "KUBE-ROUTER-INPUT"
	kubeForwardChainName   = "KUBE-ROUTER-FORWARD"
	kubeOutputChainName    = "KUBE-ROUTER-OUTPUT"
	kubeDefaultNetpolChain = "KUBE-NWPLCY-DEFAULT"
)

var (
	defaultChains = map[iptables.Chain]iptables.Chain{
		iptables.ChainInput:   kubeInputChainName,
		iptables.ChainForward: kubeForwardChainName,
		iptables.ChainOutput:  kubeOutputChainName,
	}
)

type NetworkPolicyController struct {
	ipsetMutex *sync.Mutex

	networkPolicyLister cache.Indexer
	podLister           cache.Indexer
	namespaceLister     cache.Indexer

	iptablesCmdHandler iptables.Interface
	ipsetCmdHandler    *ipset.IPSet

	serviceClusterIPRange   net.IPNet
	serviceExternalIPRanges []net.IPNet
	serviceNodePortRange    string

	filterTableRules bytes.Buffer // 不需要实例化
}

func NewNetworkPolicyController(
	networkPolicyInformer cache.SharedIndexInformer,
	podInformer cache.SharedIndexInformer,
	namespaceInformer cache.SharedIndexInformer,
	ipsetMutex *sync.Mutex,
) (*NetworkPolicyController, error) {
	ipsetCmdHandler, err := ipset.NewIPSet(false)
	if err != nil {
		return nil, err
	}
	controller := &NetworkPolicyController{
		ipsetMutex: ipsetMutex,

		iptablesCmdHandler:  iptables.New(exec.New(), iptables.ProtocolIPv4),
		ipsetCmdHandler:     ipsetCmdHandler,
		networkPolicyLister: networkPolicyInformer.GetIndexer(),
		podLister:           podInformer.GetIndexer(),
		namespaceLister:     namespaceInformer.GetIndexer(),
	}

	return controller, nil
}

// Creates custom chains KUBE-NWPLCY-DEFAULT
func (controller *NetworkPolicyController) ensureDefaultNetworkPolicyChain() {
	// `iptables -t filter -S KUBE-NWPLCY-DEFAULT 1`
	// `iptables -t filter -S KUBE-NWPLCY-DEFAULT`
	exists, err := controller.iptablesCmdHandler.ChainExists(iptables.TableFilter, kubeDefaultNetpolChain)
	if err != nil {
		klog.Fatalf("failed to check for the existence of chain %s, error: %v", kubeDefaultNetpolChain, err)
	}
	if !exists {
		// `iptables -t filter -N KUBE-NWPLCY-DEFAULT`
		_, err = controller.iptablesCmdHandler.EnsureChain(iptables.TableFilter, kubeDefaultNetpolChain)
		if err != nil {
			klog.Fatalf("failed to run iptables command to create %s chain due to %s",
				kubeDefaultNetpolChain, err.Error())
		}
	}

	// `iptables -t filter -A KUBE-NWPLCY-DEFAULT -j MARK -m comment --comment "rule to mark traffic matching a network policy" --set-xmark 0x10000/0x10000`
	markArgs := []string{"-j", "MARK", "-m", "comment", "--comment", "rule to mark traffic matching a network policy", "--set-xmark", "0x10000/0x10000"}
	_, err = controller.iptablesCmdHandler.EnsureRule(iptables.Append, iptables.TableFilter, kubeDefaultNetpolChain, markArgs...)
	if err != nil {
		klog.Fatalf("Failed to run iptables command: %s", err.Error())
	}
}

// Creates custom chains KUBE-ROUTER-INPUT, KUBE-ROUTER-FORWARD, KUBE-ROUTER-OUTPUT
// and following rules in the filter table to jump from builtin chain to custom chain
// -A INPUT   -m comment --comment "kube-router netpol" -j KUBE-ROUTER-INPUT
// -A FORWARD -m comment --comment "kube-router netpol" -j KUBE-ROUTER-FORWARD
// -A OUTPUT  -m comment --comment "kube-router netpol" -j KUBE-ROUTER-OUTPUT
func (controller *NetworkPolicyController) ensureTopLevelChains() {
	const serviceVIPPosition = "1"
	const whitelistTCPNodePortsPosition = "2"
	const whitelistUDPNodePortsPosition = "3"
	const externalIPPositionAdditive = 4

	addUUIDForRuleSpec := func(chain iptables.Chain, ruleSpec *[]string) (string, error) {
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s", chain, strings.Join(*ruleSpec, ""))))
		encoded := base32.StdEncoding.EncodeToString(hash[:])[:16] // hash rule
		for idx, part := range *ruleSpec {
			if part == "--comment" {
				(*ruleSpec)[idx+1] = (*ruleSpec)[idx+1] + " - " + encoded
				return encoded, nil
			}
		}
		return "", fmt.Errorf("could not find a comment in the ruleSpec string given: %s",
			strings.Join(*ruleSpec, " "))
	}

	for builtinChain, customChain := range defaultChains {
		_, err := controller.iptablesCmdHandler.EnsureChain(iptables.TableFilter, customChain)
		if err != nil {
			klog.Fatalf("failed to run iptables command to create %s chain due to %s", customChain, err.Error())
		}

		args := []string{"-m", "comment", "--comment", "kube-router netpol", "-j", string(customChain)}
		_, err = addUUIDForRuleSpec(builtinChain, &args)
		if err != nil {
			klog.Fatalf("Failed to get uuid for rule: %s", err.Error())
		}

		// iptables -t filter -I INPUT -m comment --comment "kube-router netpol-xxx" -j KUBE-ROUTER-INPUT
		// iptables -t filter -I FORWARD -m comment --comment "kube-router netpol-xxx" -j KUBE-ROUTER-FORWARD
		// iptables -t filter -I OUTPUT -m comment --comment "kube-router netpol-xxx" -j KUBE-ROUTER-OUTPUT
		_, err = controller.iptablesCmdHandler.EnsureRule(iptables.Prepend, iptables.TableFilter, builtinChain, args...)
		if err != nil {
			klog.Fatalf("Failed to run iptables command to insert in %s chain %s", builtinChain, err.Error())
			return
		}
	}

	// iptables -t filter -I KUBE-ROUTER-INPUT 1 -m comment --comment "allow traffic to cluster IP" -d "10.20.30.0/24" -j RETURN
	whitelistServiceVipsArgs := []string{serviceVIPPosition, "-m", "comment", "--comment", "allow traffic to cluster IP", "-d",
		controller.serviceClusterIPRange.String(), "-j", "RETURN"}
	_, err := addUUIDForRuleSpec(kubeInputChainName, &whitelistServiceVipsArgs)
	if err != nil {
		klog.Fatalf("Failed to get uuid for rule: %s", err.Error())
	}
	_, err = controller.iptablesCmdHandler.EnsureRule(iptables.Prepend, iptables.TableFilter, kubeInputChainName, whitelistServiceVipsArgs...)
	if err != nil {
		klog.Fatalf("Failed to run iptables command to insert in %s chain %s", kubeInputChainName, err.Error())
		return
	}

	// iptables -t filter -I KUBE-ROUTER-INPUT 2 -p tcp -m comment --comment "allow LOCAL UDP traffic to node ports-xxx" -m addrtype --dst-type LOCAL -m multiport --dports "30000-32767" -j RETURN
	whitelistTCPNodeportsArgs := []string{whitelistTCPNodePortsPosition, "-p", "tcp", "-m", "comment", "--comment",
		"allow LOCAL TCP traffic to node ports", "-m", "addrtype", "--dst-type", "LOCAL",
		"-m", "multiport", "--dports", controller.serviceNodePortRange, "-j", "RETURN"}
	_, err = addUUIDForRuleSpec(kubeInputChainName, &whitelistTCPNodeportsArgs)
	if err != nil {
		klog.Fatalf("Failed to get uuid for rule: %s", err.Error())
	}
	_, err = controller.iptablesCmdHandler.EnsureRule(iptables.Prepend, iptables.TableFilter, kubeInputChainName, whitelistTCPNodeportsArgs...)
	if err != nil {
		klog.Fatalf("Failed to run iptables command to insert in %s chain %s", kubeInputChainName, err.Error())
		return
	}

	// iptables -t filter -I KUBE-ROUTER-INPUT 3 -p udp -m comment --comment "allow LOCAL UDP traffic to node ports-xxx" -m addrtype --dst-type LOCAL -m multiport --dports "30000-32767" -j RETURN
	whitelistUDPNodeportsArgs := []string{whitelistUDPNodePortsPosition, "-p", "udp", "-m", "comment", "--comment",
		"allow LOCAL UDP traffic to node ports", "-m", "addrtype", "--dst-type", "LOCAL",
		"-m", "multiport", "--dports", controller.serviceNodePortRange, "-j", "RETURN"}
	_, err = addUUIDForRuleSpec(kubeInputChainName, &whitelistUDPNodeportsArgs)
	if err != nil {
		klog.Fatalf("Failed to get uuid for rule: %s", err.Error())
	}
	_, err = controller.iptablesCmdHandler.EnsureRule(iptables.Prepend, iptables.TableFilter, kubeInputChainName, whitelistUDPNodeportsArgs...)
	if err != nil {
		klog.Fatalf("Failed to run iptables command to insert in %s chain %s", kubeInputChainName, err.Error())
		return
	}

	// iptables -t filter -I KUBE-ROUTER-INPUT 4 -m comment --comment "allow traffic to external IP range:10.20.30.0/24-xxx" -d 10.20.30.0/24 -j RETURN
	for externalIPIndex, externalIPRange := range controller.serviceExternalIPRanges {
		position := fmt.Sprintf("%d", externalIPIndex+externalIPPositionAdditive)
		whitelistServiceExternalIPArgs := []string{position, "-m", "comment", "--comment",
			"allow traffic to external IP range: " + externalIPRange.String(), "-d", externalIPRange.String(),
			"-j", "RETURN"}
		_, err = addUUIDForRuleSpec(kubeInputChainName, &whitelistServiceExternalIPArgs)
		if err != nil {
			klog.Fatalf("Failed to get uuid for rule: %s", err.Error())
		}
		_, err = controller.iptablesCmdHandler.EnsureRule(iptables.Prepend, iptables.TableFilter, kubeInputChainName, whitelistServiceExternalIPArgs...)
		if err != nil {
			klog.Fatalf("Failed to run iptables command to insert in %s chain %s", kubeInputChainName, err.Error())
			return
		}
	}
}

func (controller *NetworkPolicyController) Run() {
	t := time.NewTicker(controller.syncPeriod)
	defer t.Stop()

	// setup kube-router specific top level custom chains (KUBE-ROUTER-INPUT, KUBE-ROUTER-FORWARD, KUBE-ROUTER-OUTPUT)
	controller.ensureTopLevelChains()

	// setup default network policy chain that is applied to traffic from/to the pods that does not match any network policy
	controller.ensureDefaultNetworkPolicyChain()

	for {
		select {
		case <-controller.stopCh:

		case <-t.C:

		}
	}

}

// Sync synchronizes iptables to desired state of network policies
func (controller *NetworkPolicyController) syncPolicy() {

	// setup kube-router specific top level custom chains (KUBE-ROUTER-INPUT, KUBE-ROUTER-FORWARD, KUBE-ROUTER-OUTPUT)
	controller.ensureTopLevelChains()

	// setup default network policy chain that is applied to traffic from/to the pods that does not match any network policy
	controller.ensureDefaultNetworkPolicyChain()

	controller.filterTableRules.Reset()
	err := controller.iptablesCmdHandler.SaveInto(iptables.TableFilter, &controller.filterTableRules)
	if err != nil {
		klog.Errorf("Aborting sync. Failed to run iptables-save: %v" + err.Error())
		return
	}

	networkPoliciesInfo, err := controller.buildNetworkPoliciesInfo()
	if err != nil {
		klog.Errorf("Aborting sync. Failed to build network policies: %v", err.Error())
		return
	}
	activePolicyChains, activePolicyIPSets, err := controller.syncNetworkPolicyChains(networkPoliciesInfo)
	if err != nil {
		klog.Errorf("Aborting sync. Failed to sync network policy chains: %v" + err.Error())
		return
	}
	activePodFwChains := controller.syncPodFirewallChains(networkPoliciesInfo)
	// Makes sure that the ACCEPT rules for packets marked with "0x20000" are added to the end of each of kube-router's
	// top level chains
	controller.ensureExplicitAccept()
	err = controller.cleanupStaleRules(activePolicyChains, activePodFwChains, false)
	if err != nil {
		klog.Errorf("Aborting sync. Failed to cleanup stale iptables rules: %v", err.Error())
		return
	}

	err = controller.iptablesCmdHandler.Restore(iptables.TableFilter, controller.filterTableRules.Bytes(),
		iptables.FlushTables, iptables.NoRestoreCounters)
	if err != nil {
		klog.Errorf("Aborting sync. Failed to run iptables-restore: %v\n%s",
			err.Error(), controller.filterTableRules.String())
		return
	}

	err = controller.cleanupStaleIPSets(activePolicyIPSets)
	if err != nil {
		klog.Errorf("Failed to cleanup stale ipsets: %v", err.Error())
		return
	}

}
