package controller

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/networking"
	"strings"
)

func (controller *NetworkPolicyController) syncPodFirewallChains(networkPoliciesInfo []networkPolicyInfo) map[string]bool {
	activePodFwChains := make(map[string]bool)

	// loop through the pods running on the node
	allLocalPods := controller.getLocalPods(controller.nodeIP.String())
	for _, pod := range allLocalPods {
		// ensure pod specific firewall chain exist for all the pods that need ingress firewall
		podFwChainName := podFirewallChainName(pod.namespace, pod.name)
		controller.filterTableRules.WriteString(":" + podFwChainName + "\n")
		activePodFwChains[podFwChainName] = true

		// setup rules to run through applicable ingress/egress network policies for the pod
		controller.setupPodNetpolRules(pod, podFwChainName, networkPoliciesInfo)

		// setup rules to intercept inbound traffic to the pods
		controller.interceptPodInboundTraffic(pod, podFwChainName)

		// setup rules to intercept inbound traffic to the pods
		controller.interceptPodOutboundTraffic(pod, podFwChainName)

		controller.dropUnmarkedTraffic(pod, podFwChainName)

		// set mark to indicate traffic from/to the pod passed network policies.
		// Mark will be checked to explicitly ACCEPT the traffic
		args := []string{"-A", podFwChainName, "-m", "comment", "--comment", "set mark to ACCEPT traffic that comply to network policies",
			"-j", "MARK", "--set-mark", "0x20000/0x20000", "\n"}
		controller.filterTableRules.WriteString(strings.Join(args, " "))
	}

	return activePodFwChains
}

func (controller *NetworkPolicyController) getLocalPods(nodeIP string) map[string]podInfo {
	localPods := make(map[string]podInfo)
	for _, obj := range controller.podLister.List() {
		pod := obj.(*corev1.Pod)
		if !isNetworkPolicyPod(pod) || pod.Status.HostIP != nodeIP {
			continue
		}

		localPods[pod.Status.PodIP] = podInfo{ip: pod.Status.PodIP,
			name:      pod.ObjectMeta.Name,
			namespace: pod.ObjectMeta.Namespace,
			labels:    pod.ObjectMeta.Labels,
		}
	}

	return localPods
}

// setup rules to jump to applicable network policy chains for the traffic from/to the pod
func (controller *NetworkPolicyController) setupPodNetpolRules(pod podInfo, podFwChainName string, networkPoliciesInfo []networkPolicyInfo) {
	hasIngressPolicy := false
	hasEgressPolicy := false

	// add entries in pod firewall to run through applicable network policies
	for _, policy := range networkPoliciesInfo {
		if _, ok := policy.targetPods[pod.ip]; !ok {
			continue
		}

		policyChainName := networkPolicyChainName(policy.namespace, policy.name)
		var args []string
		comment := fmt.Sprintf("run through nw policy %s", policy.name)
		if len(policy.policyTypes) == 1 {
			switch policy.policyTypes[0] {
			case networking.PolicyTypeIngress:
				hasIngressPolicy = true
				args = []string{"-I", podFwChainName, "1", "-d", pod.ip, "-m", "comment", "--comment", comment,
					"-j", policyChainName, "\n"}
			case networking.PolicyTypeEgress:
				hasEgressPolicy = true
				args = []string{"-I", podFwChainName, "1", "-s", pod.ip, "-m", "comment", "--comment", comment,
					"-j", policyChainName, "\n"}
			}
		} else if len(policy.policyTypes) == 2 {
			if (policy.policyTypes[0] == networking.PolicyTypeIngress && policy.policyTypes[1] == networking.PolicyTypeEgress) ||
				(policy.policyTypes[1] == networking.PolicyTypeIngress && policy.policyTypes[0] == networking.PolicyTypeEgress) {
				hasIngressPolicy = true
				hasEgressPolicy = true
				args = []string{"-I", podFwChainName, "1", "-m", "comment", "--comment", comment,
					"-j", policyChainName, "\n"}
			}
		}

		// INFO: `iptables -I {podChainName} 1 [-d {podIP}/-s {podIP}] -m comment --comment {comment} -j {policyChainName}`
		controller.filterTableRules.WriteString(strings.Join(args, " "))
	}

	// if pod does not have any network policy which applies rules for pod's ingress traffic
	// then apply default network policy
	if !hasIngressPolicy {
		args := []string{"-I", podFwChainName, "1", "-d", pod.ip, "-m", "comment", "--comment", "run through default ingress network policy  chain",
			"-j", kubeDefaultNetpolChain, "\n"}
		controller.filterTableRules.WriteString(strings.Join(args, " "))
	}

	// if pod does not have any network policy which applies rules for pod's egress traffic
	// then apply default network policy
	if !hasEgressPolicy {
		args := []string{"-I", podFwChainName, "1", "-s", pod.ip, "-m", "comment", "--comment", "run through default egress network policy chain",
			"-j", kubeDefaultNetpolChain, "\n"}
		controller.filterTableRules.WriteString(strings.Join(args, " "))
	}

	// INFO: This module matches packets based on their address type. Address types are used within the kernel networking
	//  stack and categorize addresses into various groups. The exact definition of that group depends on the specific layer three protocol.
	//  UNICAST LOCAL BROADCAST ANYCAST MULTICAST BLACKHOLE
	args := []string{"-I", podFwChainName, "1", "-m", "comment", "--comment", "rule to permit the traffic to pods when source is the pod's local node",
		"-m", "addrtype", "--src-type", "LOCAL", "-d", pod.ip, "-j", "ACCEPT", "\n"}
	controller.filterTableRules.WriteString(strings.Join(args, " "))

	// ensure statefull firewall drops INVALID state traffic from/to the pod
	// For full context see: https://bugzilla.netfilter.org/show_bug.cgi?id=693
	// The NAT engine ignores any packet with state INVALID, because there's no reliable way to determine what kind of
	// NAT should be performed. So the proper way to prevent the leakage is to drop INVALID packets.
	// In the future, if we ever allow services or nodes to disable conntrack checking, we may need to make this
	// conditional so that non-tracked traffic doesn't get dropped as invalid.
	args = []string{"-I", podFwChainName, "1", "-m", "comment", "--comment", "rule to drop invalid state for pod",
		"-m", "conntrack", "--ctstate", "INVALID", "-j", "DROP", "\n"}
	controller.filterTableRules.WriteString(strings.Join(args, " "))

	// ensure statefull firewall that permits RELATED,ESTABLISHED traffic from/to the pod
	args = []string{"-I", podFwChainName, "1", "-m", "comment", "--comment", "rule for stateful firewall for pod",
		"-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT", "\n"}
	controller.filterTableRules.WriteString(strings.Join(args, " "))
}

func (controller *NetworkPolicyController) interceptPodInboundTraffic(pod podInfo, podFwChainName string) {
	// ensure there is rule in filter table and FORWARD chain to jump to pod specific firewall chain
	// this rule applies to the traffic getting routed (coming for other node pods)
	comment := fmt.Sprintf("rule to jump traffic destined to POD name:%s namespace:%s to chain: %s", pod.name, pod.namespace, podFwChainName)
	args := []string{"-A", kubeForwardChainName, "-m", "comment", "--comment", comment, "-d", pod.ip, "-j", podFwChainName + "\n"}
	controller.filterTableRules.WriteString(strings.Join(args, " "))

	// ensure there is rule in filter table and OUTPUT chain to jump to pod specific firewall chain
	// this rule applies to the traffic from a pod getting routed back to another pod on same node by service proxy
	args = []string{"-A", kubeOutputChainName, "-m", "comment", "--comment", comment, "-d", pod.ip, "-j", podFwChainName + "\n"}
	controller.filterTableRules.WriteString(strings.Join(args, " "))

	// ensure there is rule in filter table and forward chain to jump to pod specific firewall chain
	// this rule applies to the traffic getting switched (coming for same node pods)
	args = []string{"-A", kubeForwardChainName, "-m", "physdev", "--physdev-is-bridged",
		"-m", "comment", "--comment", comment, "-d", pod.ip, "-j", podFwChainName, "\n"}
	controller.filterTableRules.WriteString(strings.Join(args, " "))
}

// setup iptable rules to intercept outbound traffic from pods and run it across the
// firewall chain corresponding to the pod so that egress network policies are enforced
func (controller *NetworkPolicyController) interceptPodOutboundTraffic(pod podInfo, podFwChainName string) {
	comment := fmt.Sprintf("rule to jump traffic destined to POD name:%s namespace:%s to chain: %s", pod.name, pod.namespace, podFwChainName)
	for _, chain := range defaultChains {
		// ensure there is rule in filter table and FORWARD chain to jump to pod specific firewall chain
		// this rule applies to the traffic getting forwarded/routed (traffic from the pod destined
		// to pod on a different node)
		args := []string{"-A", string(chain), "-m", "comment", "--comment", comment, "-s", pod.ip, "-j", podFwChainName, "\n"}
		controller.filterTableRules.WriteString(strings.Join(args, " "))
	}

	// ensure there is rule in filter table and forward chain to jump to pod specific firewall chain
	// this rule applies to the traffic getting switched (coming for same node pods)
	args := []string{"-A", kubeForwardChainName, "-m", "physdev", "--physdev-is-bridged",
		"-m", "comment", "--comment", comment, "-s", pod.ip, "-j", podFwChainName, "\n"}
	controller.filterTableRules.WriteString(strings.Join(args, " "))
}

func (controller *NetworkPolicyController) dropUnmarkedTraffic(pod podInfo, podFwChainName string) {
	// add rule to log the packets that will be dropped due to network policy enforcement
	comment := fmt.Sprintf("rule to log dropped traffic POD name:%s namespace:%s", pod.name, pod.namespace)
	args := []string{"-A", podFwChainName, "-m", "comment", "--comment", comment, "-m", "mark", "!", "--mark", ConntrackMark,
		"-j", "NFLOG", "--nflog-group", "100", "-m", "limit", "--limit", "10/minute", "--limit-burst", "10", "\n"}
	// This used to be AppendUnique when we were using iptables directly, this checks to make sure we didn't drop
	// unmarked for this chain already
	if strings.Contains(controller.filterTableRules.String(), strings.Join(args, " ")) {
		return
	}
	controller.filterTableRules.WriteString(strings.Join(args, " "))

	// add rule to DROP if no applicable network policy permits the traffic
	comment = fmt.Sprintf("rule to REJECT traffic destined for POD name:%s namespace:%s", pod.name, pod.namespace)
	args = []string{"-A", podFwChainName, "-m", "comment", "--comment", comment, "-m", "mark", "!", "--mark", ConntrackMark, "-j", "REJECT", "\n"}
	controller.filterTableRules.WriteString(strings.Join(args, " "))

	// reset mark to let traffic pass through rest of the chains
	args = []string{"-A", podFwChainName, "-j", "MARK", "--set-mark", "0/0x10000", "\n"}
	controller.filterTableRules.WriteString(strings.Join(args, " "))
}

func isNetworkPolicyPod(pod *corev1.Pod) bool {
	return len(pod.Status.PodIP) != 0 && !pod.Spec.HostNetwork && pod.Status.Phase == corev1.PodRunning
}

func listPodIPBlock(peer networking.NetworkPolicyPeer) [][]string {
	ipBlock := make([][]string, 0)
	if peer.PodSelector == nil && peer.NamespaceSelector == nil && peer.IPBlock != nil {
		ipBlock = append(ipBlock, []string{peer.IPBlock.CIDR, "timeout", "0"})
		for _, except := range peer.IPBlock.Except {
			ipBlock = append(ipBlock, []string{except, "timeout", "0", "nomatch"})
		}
	}

	return ipBlock
}
