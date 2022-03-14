package controller

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"k8s.io/klog/v2"
	"strings"
	"time"

	"k8s.io/kubernetes/pkg/util/ipset"
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
		"INPUT":   kubeInputChainName,
		"FORWARD": kubeForwardChainName,
		"OUTPUT":  kubeOutputChainName,
	}
)

type NetworkPolicyController struct {
	iptablesCmdHandler iptables.Interface
	ipsetCmdHandler    ipset.Interface
}

func NewNetworkPolicyController() *NetworkPolicyController {
	controller := &NetworkPolicyController{
		iptablesCmdHandler: iptables.New(exec.New(), iptables.ProtocolIPv4),
		ipsetCmdHandler:    ipset.New(exec.New()),
	}

	return controller
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

	addUUIDForRuleSpec := func(chain string, ruleSpec *[]string) (string, error) {
		hash := sha256.Sum256([]byte(chain + strings.Join(*ruleSpec, "")))
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

		_, err = controller.iptablesCmdHandler.EnsureRule(iptables.Prepend, iptables.TableFilter, builtinChain, args...)
		if err != nil {
			klog.Fatalf("Failed to run iptables command to insert in %s chain %s", builtinChain, err.Error())
			return
		}
	}

	whitelistServiceVipsArgs := []string{serviceVIPPosition, "-m", "comment", "--comment", "allow traffic to cluster IP", "-d",
		controller.serviceClusterIPRange.String(), "-j", "RETURN"}
	uuid, err := addUUIDForRuleSpec(kubeInputChainName, &whitelistServiceVipsArgs)
	if err != nil {
		klog.Fatalf("Failed to get uuid for rule: %s", err.Error())
	}
	_, err = controller.iptablesCmdHandler.EnsureRule(iptables.Prepend, iptables.TableFilter, kubeInputChainName, whitelistServiceVipsArgs...)
	if err != nil {
		klog.Fatalf("Failed to run iptables command to insert in %s chain %s", kubeInputChainName, err.Error())
		return
	}

	whitelistTCPNodeportsArgs := []string{whitelistTCPNodePortsPosition, "-p", "tcp", "-m", "comment", "--comment",
		"allow LOCAL TCP traffic to node ports", "-m", "addrtype", "--dst-type", "LOCAL",
		"-m", "multiport", "--dports", controller.serviceNodePortRange, "-j", "RETURN"}
	uuid, err = addUUIDForRuleSpec(kubeInputChainName, &whitelistTCPNodeportsArgs)
	if err != nil {
		klog.Fatalf("Failed to get uuid for rule: %s", err.Error())
	}
	_, err = controller.iptablesCmdHandler.EnsureRule(iptables.Prepend, iptables.TableFilter, kubeInputChainName, whitelistTCPNodeportsArgs...)
	if err != nil {
		klog.Fatalf("Failed to run iptables command to insert in %s chain %s", kubeInputChainName, err.Error())
		return
	}

	whitelistUDPNodeportsArgs := []string{whitelistUDPNodePortsPosition, "-p", "udp", "-m", "comment", "--comment",
		"allow LOCAL UDP traffic to node ports", "-m", "addrtype", "--dst-type", "LOCAL",
		"-m", "multiport", "--dports", controller.serviceNodePortRange, "-j", "RETURN"}
	uuid, err = addUUIDForRuleSpec(kubeInputChainName, &whitelistUDPNodeportsArgs)
	if err != nil {
		klog.Fatalf("Failed to get uuid for rule: %s", err.Error())
	}
	_, err = controller.iptablesCmdHandler.EnsureRule(iptables.Prepend, iptables.TableFilter, kubeInputChainName, whitelistUDPNodeportsArgs...)
	if err != nil {
		klog.Fatalf("Failed to run iptables command to insert in %s chain %s", kubeInputChainName, err.Error())
		return
	}

	for externalIPIndex, externalIPRange := range controller.serviceExternalIPRanges {
		position := fmt.Sprintf("%s", externalIPIndex+externalIPPositionAdditive)
		whitelistServiceExternalIPArgs := []string{position, "-m", "comment", "--comment",
			"allow traffic to external IP range: " + externalIPRange.String(), "-d", externalIPRange.String(),
			"-j", "RETURN"}
		uuid, err = addUUIDForRuleSpec(kubeInputChainName, &whitelistServiceExternalIPArgs)
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

}
