package controller

import (
	"fmt"
	"k8s-lx1036/k8s/network/network-policy/pkg/ipset"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/apis/networking"
	"strings"
)

// internal structure to represent a network policy
type networkPolicyInfo struct {
	name        string
	namespace   string
	podSelector labels.Selector

	// set of pods matching network policy spec podselector label selector
	targetPods map[string]podInfo

	// whitelist ingress rules from the network policy spec
	ingressRules []ingressRule

	// whitelist egress rules from the network policy spec
	egressRules []egressRule

	// policy type "ingress" or "egress" or "both" as defined by PolicyType in the spec
	policyTypes []networking.PolicyType
}

type podInfo struct {
	ip        string
	name      string
	namespace string
	labels    map[string]string
}
type ingressRule struct {
	matchAllPorts  bool
	ports          []protocolAndPort
	namedPorts     []endPoints
	matchAllSource bool
	srcPods        []podInfo
	srcIPBlocks    [][]string
}
type egressRule struct {
	matchAllPorts        bool
	ports                []protocolAndPort
	namedPorts           []endPoints
	matchAllDestinations bool
	dstPods              []podInfo
	dstIPBlocks          [][]string
}
type endPoints struct {
	ips []string
	protocolAndPort
}
type protocolAndPort struct {
	protocol string
	port     string
	endport  string
}

func (controller *NetworkPolicyController) buildNetworkPoliciesInfo() ([]networkPolicyInfo, error) {
	var networkPolicyInfos []networkPolicyInfo
	for _, item := range controller.networkPolicyLister.List() {
		networkPolicy := item.(*networking.NetworkPolicy)
		podSelector, _ := metav1.LabelSelectorAsSelector(&networkPolicy.Spec.PodSelector)
		policyInfo := networkPolicyInfo{
			name:        networkPolicy.Name,
			namespace:   networkPolicy.Namespace,
			podSelector: podSelector,
			policyTypes: networkPolicy.Spec.PolicyTypes,
		}

		// target pods
		pods, err := listerv1.NewPodLister(controller.podLister).Pods(networkPolicy.Namespace).List(podSelector)
		if err == nil {
			for _, pod := range pods {
				if !isNetworkPolicyPod(pod) {
					continue
				}
				policyInfo.targetPods[pod.Status.PodIP] = podInfo{
					ip:        pod.Status.PodIP,
					name:      pod.ObjectMeta.Name,
					namespace: pod.ObjectMeta.Namespace,
					labels:    pod.ObjectMeta.Labels,
				}
				controller.grabNamedPortFromPod(pod, &namedPort2IngressEps)
			}
		}

		policyInfo.ingressRules = append(policyInfo.ingressRules, controller.buildNetworkPolicyIngressRule(networkPolicy)...)
		policyInfo.egressRules = append(policyInfo.egressRules, controller.buildNetworkPolicyEgressRule(networkPolicy)...)
		networkPolicyInfos = append(networkPolicyInfos, policyInfo)
	}

	return networkPolicyInfos, nil
}

func (controller *NetworkPolicyController) buildNetworkPolicyIngressRule(networkPolicy *networking.NetworkPolicy) []ingressRule {
	var ingressRules []ingressRule
	for _, ingress := range networkPolicy.Spec.Ingress {
		var rule ingressRule
		// If this field is empty or missing in the spec, this rule matches all sources
		if len(ingress.From) == 0 {
			rule.matchAllSource = true
		} else {
			rule.matchAllSource = false
			for _, peer := range ingress.From {
				if peerPods, err := controller.listPodsByNetworkPolicyPeer(peer, networkPolicy.Namespace); err == nil {
					for _, pod := range peerPods {
						if !isNetworkPolicyPod(pod) {
							continue
						}
						rule.srcPods = append(rule.srcPods, podInfo{
							ip:        pod.Status.PodIP,
							name:      pod.ObjectMeta.Name,
							namespace: pod.ObjectMeta.Namespace,
							labels:    pod.ObjectMeta.Labels,
						})
					}
				}

				rule.srcIPBlocks = append(rule.srcIPBlocks, listPodIPBlock(peer)...)
			}
		}

		if len(ingress.Ports) == 0 {
			rule.matchAllPorts = true
		} else {
			rule.matchAllPorts = false
			rule.ports, rule.namedPorts = controller.processNetworkPolicyPorts(ingress.Ports, namedPort2IngressEps)
		}

		ingressRules = append(ingressRules, rule)
	}

	return ingressRules
}

func (controller *NetworkPolicyController) buildNetworkPolicyEgressRule(networkPolicy *networking.NetworkPolicy) []egressRule {
	var egressRules []egressRule
	for _, egress := range networkPolicy.Spec.Egress {
		var rule egressRule
		// If this field is empty or missing in the spec, this rule matches all sources
		if len(egress.To) == 0 {
			rule.matchAllDestinations = true
			// if rule.To is empty but rule.Ports not, we must try to grab NamedPort from pods that in same
			// namespace, so that we can design iptables rule to describe "match all dst but match some named
			// dst-port" egress rule
			if policyRulePortsHasNamedPort(egress.Ports) {
				pods, _ := listerv1.NewPodLister(controller.podLister).Pods(networkPolicy.Namespace).List(labels.Everything())
				for _, pod := range pods {
					if !isNetworkPolicyPod(pod) {
						continue
					}
					controller.grabNamedPortFromPod(pod, &namedPort2EgressEps)
				}
			}
		} else {
			rule.matchAllDestinations = false
			for _, peer := range egress.To {
				if peerPods, err := controller.listPodsByNetworkPolicyPeer(peer, networkPolicy.Namespace); err == nil {
					for _, pod := range peerPods {
						if !isNetworkPolicyPod(pod) {
							continue
						}
						rule.dstPods = append(rule.dstPods, podInfo{
							ip:        pod.Status.PodIP,
							name:      pod.ObjectMeta.Name,
							namespace: pod.ObjectMeta.Namespace,
							labels:    pod.ObjectMeta.Labels,
						})

						controller.grabNamedPortFromPod(pod, &namedPort2EgressEps)
					}
				}

				rule.dstIPBlocks = append(rule.dstIPBlocks, listPodIPBlock(peer)...)
			}
		}

		if len(egress.Ports) == 0 {
			rule.matchAllPorts = true
		} else {
			rule.matchAllPorts = false
			rule.ports, rule.namedPorts = controller.processNetworkPolicyPorts(egress.Ports, namedPort2EgressEps)
		}

		egressRules = append(egressRules, rule)
	}

	return egressRules
}

func (controller *NetworkPolicyController) listPodsByNetworkPolicyPeer(peer networking.NetworkPolicyPeer, policyNamespace string) ([]*corev1.Pod, error) {
	var matchingPods []*corev1.Pod
	var err error
	if peer.NamespaceSelector != nil {
		namespaceSelector, _ := metav1.LabelSelectorAsSelector(peer.NamespaceSelector)
		namespaces, err := listerv1.NewNamespaceLister(controller.namespaceLister).List(namespaceSelector)
		if err != nil {
			klog.Errorf(fmt.Sprintf("failed to list namespaces in label selector:%s for err:%v", namespaceSelector.String(), err))
			return nil, err
		}
		podSelector := labels.Everything()
		if peer.PodSelector != nil {
			podSelector, _ = metav1.LabelSelectorAsSelector(peer.PodSelector)
		}
		for _, namespace := range namespaces {
			pods, err := listerv1.NewPodLister(controller.podLister).Pods(namespace.Name).List(podSelector)
			if err != nil {
				klog.Errorf(fmt.Sprintf("failed to list pods in namespace:%s for err:%v", namespace.Name, err))
				continue
			}
			matchingPods = append(matchingPods, pods...)
		}
	} else if peer.PodSelector != nil {
		podSelector, _ := metav1.LabelSelectorAsSelector(peer.PodSelector)
		matchingPods, err = listerv1.NewPodLister(controller.podLister).Pods(policyNamespace).List(podSelector)
	}

	return matchingPods, err
}

func (controller *NetworkPolicyController) processNetworkPolicyPorts(npPorts []networking.NetworkPolicyPort, namedPort2eps namedPort2eps) (numericPorts []protocolAndPort, namedPorts []endPoints) {

}

// Configure iptables rules representing each network policy. All pod's matched by
// network policy spec podselector labels are grouped together in one ipset which
// is used for matching destination ip address. Each ingress rule in the network
// policyspec is evaluated to set of matching pods, which are grouped in to a
// ipset used for source ip addr matching.
func (controller *NetworkPolicyController) syncNetworkPolicyChains(networkPoliciesInfo []networkPolicyInfo) (map[string]bool, map[string]bool, error) {
	var err error
	activePolicyChains := make(map[string]bool)
	activePolicyIPSets := make(map[string]bool)

	controller.ipsetMutex.Lock()
	defer controller.ipsetMutex.Unlock()

	// save ipset into stdout buffer
	if err := controller.ipsetCmdHandler.Save(); err != nil {
		return nil, nil, err
	}

	// run through all network policies
	for _, policy := range networkPoliciesInfo {
		// ensure there is a unique chain per network policy in filter table
		policyChainName := networkPolicyChainName(policy.namespace, policy.name)
		controller.filterTableRules.WriteString(":" + policyChainName + "\n")
		activePolicyChains[policyChainName] = true

		currentPodIPs := make([]string, 0, len(policy.targetPods))
		for ip := range policy.targetPods {
			currentPodIPs = append(currentPodIPs, ip)
		}

		for _, policyType := range policy.policyTypes {
			switch policyType {
			case networking.PolicyTypeIngress:
				// create a ipset for all destination pod ip's matched by the policy spec PodSelector
				targetDstPodIPSetName := policyDstPodIPSetName(policy.namespace, policy.name)
				controller.refreshIPSet(targetDstPodIPSetName, ipset.TypeHashIP, currentPodIPs)
				err = controller.processIngressRules(policy, targetDstPodIPSetName, activePolicyIPSets)
				if err != nil {
					return nil, nil, err
				}
				activePolicyIPSets[targetDstPodIPSetName] = true
			case networking.PolicyTypeEgress:
				// create a ipset for all source pod ip's matched by the policy spec PodSelector
				targetSourcePodIPSetName := policySrcPodIPSetName(policy.namespace, policy.name)
				controller.refreshIPSet(targetSourcePodIPSetName, ipset.TypeHashIP, currentPodIPs)
				err = controller.processEgressRules(policy, targetSourcePodIPSetName, activePolicyIPSets)
				if err != nil {
					return nil, nil, err
				}
				activePolicyIPSets[targetSourcePodIPSetName] = true
			}
		}
	}

	// restore ipset from stdout buffer
	if err := controller.ipsetCmdHandler.Restore(); err != nil {
		return nil, nil, err
	}

	return activePolicyChains, activePolicyIPSets, nil
}

func (controller *NetworkPolicyController) processIngressRules(policy networkPolicyInfo, targetDstPodIPSetName string, activePolicyIPSets map[string]bool) error {
	// From network policy spec: "If field 'Ingress' is empty then this NetworkPolicy does not allow any traffic "
	// so no whitelist rules to be added to the network policy
	if policy.ingressRules == nil {
		return nil
	}

	policyChainName := networkPolicyChainName(policy.namespace, policy.name)
	for id, rule := range policy.ingressRules {
		comment := fmt.Sprintf("rule to ACCEPT traffic from source pods to dest pods selected by policy name %s namespace %s", policy.name, policy.namespace)

		if len(rule.srcPods) != 0 {
			// Create policy based ipset with source pod IPs
			srcPodIPSetName := policyIndexedSrcPodIPSetName(policy.namespace, policy.name, id)
			activePolicyIPSets[srcPodIPSetName] = true
			controller.refreshIPSet(srcPodIPSetName, ipset.TypeHashIP, getIPsFromPods(rule.srcPods))

			// If the ingress policy contains port declarations, we need to make sure that we match on pod IP and port
			if len(rule.ports) != 0 {
				for _, port := range rule.ports {
					controller.appendIPTableRules(policyChainName, comment, srcPodIPSetName, targetDstPodIPSetName, port.protocol, port.port, port.endport)
				}
			}

			// If the ingress policy contains named port declarations, we need to make sure that we match on pod IP and
			// the resolved port number
			if len(rule.namedPorts) != 0 {
				for portIdx, eps := range rule.namedPorts {
					namedPortIPSetName := policyIndexedIngressNamedPortIPSetName(policy.namespace, policy.name, id, portIdx)
					activePolicyIPSets[namedPortIPSetName] = true
					controller.refreshIPSet(namedPortIPSetName, ipset.TypeHashIP, eps.ips)
					controller.appendIPTableRules(policyChainName, comment, srcPodIPSetName, namedPortIPSetName, eps.protocol, eps.port, eps.endport)
				}
			}

			// If the ingress policy contains no ports at all create the policy based only on IP
			if len(rule.ports) == 0 && len(rule.namedPorts) == 0 {
				// case where no 'ports' details specified in the ingress rule but 'from' details specified
				// so match on specified source and destination ip with all port and protocol
				controller.appendIPTableRules(policyChainName, comment, srcPodIPSetName, targetDstPodIPSetName, "", "", "")
			}
		}

		if rule.matchAllSource && !rule.matchAllPorts {
			for _, port := range rule.ports {
				controller.appendIPTableRules(policyChainName, comment, "", targetDstPodIPSetName, port.protocol, port.port, port.endport)
			}
			for portIdx, eps := range rule.namedPorts {
				namedPortIPSetName := policyIndexedIngressNamedPortIPSetName(policy.namespace, policy.name, id, portIdx)
				activePolicyIPSets[namedPortIPSetName] = true
				controller.refreshIPSet(namedPortIPSetName, ipset.TypeHashIP, eps.ips)
				controller.appendIPTableRules(policyChainName, comment, "", namedPortIPSetName, eps.protocol, eps.port, eps.endport)
			}
		}

		if rule.matchAllSource && rule.matchAllPorts {
			controller.appendIPTableRules(policyChainName, comment, "", targetDstPodIPSetName, "", "", "")
		}

		comment = fmt.Sprintf("rule to ACCEPT traffic from specified ipBlocks to dest pods selected by policy name %s namespace %s", policy.name, policy.namespace)
		if len(rule.srcIPBlocks) != 0 {
			srcIPBlockIPSetName := policyIndexedSourceIPBlockIPSetName(policy.namespace, policy.name, id)
			activePolicyIPSets[srcIPBlockIPSetName] = true
			controller.ipsetCmdHandler.RefreshSet(srcIPBlockIPSetName, rule.srcIPBlocks, ipset.TypeHashNet)
			if rule.matchAllPorts {
				controller.appendIPTableRules(policyChainName, comment, srcIPBlockIPSetName, targetDstPodIPSetName, "", "", "")
			} else {
				for _, port := range rule.ports {
					controller.appendIPTableRules(policyChainName, comment, srcIPBlockIPSetName, targetDstPodIPSetName, port.protocol, port.port, port.endport)
				}
				for portIdx, eps := range rule.namedPorts {
					namedPortIPSetName := policyIndexedIngressNamedPortIPSetName(policy.namespace, policy.name, id, portIdx)
					activePolicyIPSets[namedPortIPSetName] = true
					controller.refreshIPSet(namedPortIPSetName, ipset.TypeHashNet, eps.ips)
					controller.appendIPTableRules(policyChainName, comment, srcIPBlockIPSetName, namedPortIPSetName, eps.protocol, eps.port, eps.endport)
				}
			}
		}
	}

	return nil
}

func (controller *NetworkPolicyController) processEgressRules(policy networkPolicyInfo, targetSrcPodIPSetName string, activePolicyIPSets map[string]bool) error {
	if policy.egressRules == nil {
		return nil
	}

	policyChainName := networkPolicyChainName(policy.namespace, policy.name)
	for id, rule := range policy.egressRules {
		comment := fmt.Sprintf("rule to ACCEPT traffic from source pods to dest pods selected by policy name %s namespace %s", policy.name, policy.namespace)

		if len(rule.dstPods) != 0 {
			dstPodIPSetName := policyIndexedDstPodIPSetName(policy.namespace, policy.name, id)
			activePolicyIPSets[dstPodIPSetName] = true
			controller.refreshIPSet(dstPodIPSetName, ipset.TypeHashIP, getIPsFromPods(rule.dstPods))

			// If the egress policy contains port declarations, we need to make sure that we match on pod IP and port
			if len(rule.ports) != 0 {
				for _, port := range rule.ports {
					controller.appendIPTableRules(policyChainName, comment, targetSrcPodIPSetName, dstPodIPSetName, port.protocol, port.port, port.endport)
				}
			}

			// If the ingress policy contains named port declarations, we need to make sure that we match on pod IP and
			// the resolved port number
			if len(rule.namedPorts) != 0 {
				for portIdx, eps := range rule.namedPorts {
					namedPortIPSetName := policyIndexedEgressNamedPortIPSetName(policy.namespace, policy.name, id, portIdx)
					activePolicyIPSets[namedPortIPSetName] = true
					controller.refreshIPSet(namedPortIPSetName, ipset.TypeHashIP, eps.ips)
					controller.appendIPTableRules(policyChainName, comment, targetSrcPodIPSetName, namedPortIPSetName, eps.protocol, eps.port, eps.endport)
				}
			}

			// If the ingress policy contains no ports at all create the policy based only on IP
			if len(rule.ports) == 0 && len(rule.namedPorts) == 0 {
				// case where no 'ports' details specified in the ingress rule but 'from' details specified
				// so match on specified source and destination ip with all port and protocol
				controller.appendIPTableRules(policyChainName, comment, targetSrcPodIPSetName, dstPodIPSetName, "", "", "")
			}
		}

		if rule.matchAllDestinations && !rule.matchAllPorts {
			for _, port := range rule.ports {
				controller.appendIPTableRules(policyChainName, comment, targetSrcPodIPSetName, "", port.protocol, port.port, port.endport)
			}
			for _, eps := range rule.namedPorts {
				controller.appendIPTableRules(policyChainName, comment, targetSrcPodIPSetName, "", eps.protocol, eps.port, eps.endport)
			}
		}

		if rule.matchAllDestinations && rule.matchAllPorts {
			controller.appendIPTableRules(policyChainName, comment, targetSrcPodIPSetName, "", "", "", "")
		}

		comment = fmt.Sprintf("rule to ACCEPT traffic from source pods to specified ipBlocks selected by policy name %s namespace %s", policy.name, policy.namespace)
		if len(rule.dstIPBlocks) != 0 {
			dstIPBlockIPSetName := policyIndexedDstIPBlockIPSetName(policy.namespace, policy.name, id)
			activePolicyIPSets[dstIPBlockIPSetName] = true
			controller.ipsetCmdHandler.RefreshSet(dstIPBlockIPSetName, rule.dstIPBlocks, ipset.TypeHashNet)
			if rule.matchAllPorts {
				controller.appendIPTableRules(policyChainName, comment, targetSrcPodIPSetName, dstIPBlockIPSetName, "", "", "")
			} else {
				for _, port := range rule.ports {
					controller.appendIPTableRules(policyChainName, comment, targetSrcPodIPSetName, dstIPBlockIPSetName, port.protocol, port.port, port.endport)
				}
			}
		}
	}

	return nil
}

func (controller *NetworkPolicyController) refreshIPSet(ipsetName, setType string, ips []string) {
	setEntries := make([][]string, 0)
	for _, ip := range ips {
		setEntries = append(setEntries, []string{ip, ipset.OptionTimeout, "0"})
	}

	controller.ipsetCmdHandler.RefreshSet(ipsetName, setEntries, setType)
}

// https://linux.die.net/man/8/iptables --set
// `iptables -A {chain} -m comment --comment {comment} -m set --match-set {match-set} src -m set --match-set {match-set} dst -p {protocol} -dport {port} -j MARK --set-xmark 0x10000/0x10000`
// `iptables -A {chain} -m comment --comment {comment} -m set --match-set {match-set} src -m set --match-set {match-set} dst -p {protocol} -dport {port} -m mark --mark 0x10000/0x10000 -j RETURN`
func (controller *NetworkPolicyController) appendIPTableRules(policyChainName, comment, srcIPSetName, dstIPSetName,
	protocol, dPort, endPort string) {
	args := make([]string, 0)
	args = append(args, "-A", policyChainName)
	if comment != "" {
		args = append(args, "-m", "comment", "--comment", comment)
	}
	if srcIPSetName != "" {
		args = append(args, "-m", "set", "--match-set", srcIPSetName, "src")
	}
	if dstIPSetName != "" {
		args = append(args, "-m", "set", "--match-set", dstIPSetName, "dst")
	}
	if protocol != "" {
		args = append(args, "-p", protocol)
	}
	if dPort != "" {
		if endPort != "" {
			args = append(args, "--dport", fmt.Sprintf("%s:%s", dPort, endPort))
		} else {
			args = append(args, "--dport", dPort)
		}
	}

	markArgs := append(args, "-j", "MARK", "--set-xmark", "0x10000/0x10000", "\n")
	controller.filterTableRules.WriteString(strings.Join(markArgs, " "))
	args = append(args, "-m", "mark", "--mark", "0x10000/0x10000", "-j", "RETURN", "\n")
	controller.filterTableRules.WriteString(strings.Join(args, " "))
}

func policyRulePortsHasNamedPort(npPorts []networking.NetworkPolicyPort) bool {
	for _, npPort := range npPorts {
		if npPort.Port != nil && npPort.Port.Type == intstr.String {
			return true
		}
	}
	return false
}
