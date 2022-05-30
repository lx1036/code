package controller

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/loadbalancer/network-policy/pkg/ipset"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/apis/networking"
	"k8s.io/kubernetes/pkg/util/iptables"
	"k8s.io/utils/exec"
)

// INFO: @see https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/iptables/proxier.go

const (
	kubeInputChainName     = "KUBE-ROUTER-INPUT"
	kubeForwardChainName   = "KUBE-ROUTER-FORWARD"
	kubeOutputChainName    = "KUBE-ROUTER-OUTPUT"
	kubeDefaultNetpolChain = "KUBE-NWPLCY-DEFAULT"

	ConntrackMark = "0x10000/0x10000"
)

// INFO: 为何是 INPUT/OUTPUT/FORWARD 这三个 chain？因为 filter table 包含这三个 table，针对每一个内置的 chain 则 -j 自定义的 chain
var (
	iptablesJumpChains = map[iptables.Chain]iptables.Chain{
		iptables.ChainInput:   kubeInputChainName,
		iptables.ChainForward: kubeForwardChainName,
		iptables.ChainOutput:  kubeOutputChainName,
	}
)

// iptables -t filter --list-rules

type NetworkPolicyController struct {
	sync.Mutex
	ipsetMutex *sync.Mutex

	networkPolicyLister cache.Indexer
	podLister           cache.Indexer
	namespaceLister     cache.Indexer

	// RETURN means stop traversing this chain and resume at the next rule in the previous (calling) chain.
	iptablesCmdHandler iptables.Interface
	ipsetCmdHandler    *ipset.IPSet

	serviceClusterIPRange   net.IPNet
	serviceExternalIPRanges []net.IPNet
	serviceNodePortRange    string

	filterTableRules *bytes.Buffer
	syncPeriod       time.Duration
	nodeIP           net.IP

	syncChan chan struct{}
	stopCh   chan struct{}
}

func NewNetworkPolicyController(
	clientset kubernetes.Interface,
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

		iptablesCmdHandler: iptables.New(exec.New(), iptables.ProtocolIPv4),
		ipsetCmdHandler:    ipsetCmdHandler,

		filterTableRules: bytes.NewBuffer(nil),

		networkPolicyLister: networkPolicyInformer.GetIndexer(),
		podLister:           podInformer.GetIndexer(),
		namespaceLister:     namespaceInformer.GetIndexer(),

		syncChan: make(chan struct{}, 1), // network policy 同时只能处理一个，因为前一个会影响后面的. 通过 channel 可以精细化控制 sync 速度, @see sync()
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName != "" {
		node, err := clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		nodeIP, err := GetNodeIP(node)
		if err != nil {
			return nil, err
		}
		controller.nodeIP = nodeIP
	} else {
		return nil, fmt.Errorf("NODE_NAME env is empty")
	}

	networkPolicyInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			policy, ok := obj.(*networking.NetworkPolicy)
			if !ok {
				return
			}
			controller.onNetworkPolicyUpdate(policy)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			policy, ok := newObj.(*networking.NetworkPolicy)
			if !ok {
				return
			}
			controller.onNetworkPolicyUpdate(policy)
		},
		DeleteFunc: func(obj interface{}) {
			policy, ok := obj.(*networking.NetworkPolicy)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					klog.Errorf("unexpected object type: %v", obj)
					return
				}
				if policy, ok = tombstone.Obj.(*networking.NetworkPolicy); !ok {
					klog.Errorf("unexpected object type: %v", obj)
					return
				}
			}

			controller.onNetworkPolicyUpdate(policy)
		},
	})

	namespaceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ns := obj.(*corev1.Namespace)
			if ns.Labels == nil {
				return
			}
			controller.syncPolicy()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if reflect.DeepEqual(oldObj.(*corev1.Namespace).Labels, newObj.(*corev1.Namespace).Labels) {
				return
			}
			controller.syncPolicy()
		},
		DeleteFunc: func(obj interface{}) {
			ns, ok := obj.(*corev1.Namespace)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					klog.Errorf("unexpected object type: %v", obj)
					return
				}
				if ns, ok = tombstone.Obj.(*corev1.Namespace); !ok {
					klog.Errorf("unexpected object type: %v", obj)
					return
				}
			}

			if ns.Labels == nil {
				return
			}
			controller.syncPolicy()
		},
	})

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			if isNetworkPolicyPod(pod) {
				controller.syncPolicy()
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if isPodUpdateNetPolRelevant(oldObj.(*corev1.Pod), newObj.(*corev1.Pod)) {
				controller.syncPolicy()
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					klog.Errorf("unexpected object type: %v", obj)
					return
				}
				if pod, ok = tombstone.Obj.(*corev1.Pod); !ok {
					klog.Errorf("unexpected object type: %v", obj)
					return
				}
			}
			if pod.Labels == nil {
				return
			}
			controller.syncPolicy()
		},
	})

	return controller, nil
}

// Creates custom chains KUBE-NWPLCY-DEFAULT
func (controller *NetworkPolicyController) ensureDefaultNetworkPolicyChain() {
	// `iptables -t filter -S KUBE-NWPLCY-DEFAULT 1`
	// `iptables -t filter -S KUBE-NWPLCY-DEFAULT`
	// `iptables -t filter -N KUBE-NWPLCY-DEFAULT`
	if _, err := controller.iptablesCmdHandler.EnsureChain(iptables.TableFilter, kubeDefaultNetpolChain); err != nil {
		klog.Fatalf("failed to run iptables command to create %s chain due to %s",
			kubeDefaultNetpolChain, err.Error())
	}

	// `iptables -t filter -A KUBE-NWPLCY-DEFAULT -j MARK -m comment --comment "rule to mark traffic matching a network policy" --set-xmark 0x10000/0x10000`
	markArgs := []string{"-j", "MARK", "-m", "comment", "--comment", "rule to mark traffic matching a network policy", "--set-xmark", "0x10000/0x10000"}
	if _, err := controller.iptablesCmdHandler.EnsureRule(iptables.Append, iptables.TableFilter, kubeDefaultNetpolChain, markArgs...); err != nil {
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

	for builtinChain, customChain := range iptablesJumpChains {
		_, err := controller.iptablesCmdHandler.EnsureChain(iptables.TableFilter, customChain)
		if err != nil {
			klog.Fatalf("failed to run iptables command to create %s chain due to %s", customChain, err.Error())
		}

		args := []string{"-m", "comment", "--comment", `"kube-router netpol"`, "-j", string(customChain)}
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
			return

		case <-controller.syncChan:
			controller.ipsetMutex.Lock()
			controller.syncPolicy()
			controller.ipsetMutex.Unlock()

		case <-t.C:
			controller.syncPolicy()
		}
	}
}

func (controller *NetworkPolicyController) sync() {
	select {
	case controller.syncChan <- struct{}{}:
	default:
		klog.Infof("Already pending sync, dropping request")
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

func (controller *NetworkPolicyController) ensureExplicitAccept() {
	// for the traffic to/from the local pod's let network policy controller be
	// authoritative entity to ACCEPT the traffic if it complies to network policies
	for _, customChain := range iptablesJumpChains {
		args := []string{"-m", "comment", "--comment", "rule to explicitly ACCEPT traffic that comply to network policies",
			"-m", "mark", "--mark", "0x20000/0x20000", "-j", "ACCEPT"}

		controller.filterTableRules = AppendUnique(controller.filterTableRules, string(customChain), args)
	}
}

func AppendUnique(buffer bytes.Buffer, chain string, rule []string) bytes.Buffer {
	var desiredBuffer bytes.Buffer
	rules := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	for _, foundRule := range rules {
		if strings.Contains(foundRule, chain) && strings.Contains(foundRule, strings.Join(rule, " ")) {
			continue
		}
		desiredBuffer.WriteString(foundRule + "\n")
	}
	ruleStr := strings.Join(append([]string{"-A", chain}, rule...), " ")
	desiredBuffer.WriteString(ruleStr + "\n")
	return desiredBuffer
}

func (controller *NetworkPolicyController) cleanupStaleRules(activePolicyChains, activePodFwChains map[string]bool,
	deleteDefaultChains bool) error {

	return nil
}

func (controller *NetworkPolicyController) cleanupStaleIPSets(activePolicyIPSets map[string]bool) error {
	controller.ipsetMutex.Lock()
	defer controller.ipsetMutex.Unlock()

	cleanupPolicyIPSets := make([]*ipset.Set, 0)
	ipsets, err := ipset.NewIPSet(false)
	if err != nil {
		return fmt.Errorf("failed to create ipsets command executor due to %s", err.Error())
	}
	err = ipsets.Save()
	if err != nil {
		klog.Fatalf("failed to initialize ipsets command executor due to %s", err.Error())
	}
	for _, set := range ipsets.Sets {
		if strings.HasPrefix(set.Name, kubeSourceIPSetPrefix) ||
			strings.HasPrefix(set.Name, kubeDestinationIPSetPrefix) {
			if _, ok := activePolicyIPSets[set.Name]; !ok {
				cleanupPolicyIPSets = append(cleanupPolicyIPSets, set)
			}
		}
	}
	// cleanup network policy ipsets
	for _, set := range cleanupPolicyIPSets {
		err = set.Destroy()
		if err != nil {
			return fmt.Errorf("failed to delete ipset %s due to %s", set.Name, err)
		}
	}
	return nil
}

// isPodUpdateNetPolRelevant checks the attributes that we care about for building NetworkPolicies on the host and if it
// finds a relevant change, it returns true otherwise it returns false. The things we care about for NetworkPolicies:
// 1) Is the phase of the pod changing? (matters for catching completed, succeeded, or failed jobs)
// 2) Is the pod IP changing? (changes how the network policy is applied to the host)
// 3) Is the pod's host IP changing? (should be caught in the above, with the CNI kube-router runs with but we check
//     this as well for sanity)
// 4) Is a pod's label changing? (potentially changes which NetworkPolicies select this pod)
func isPodUpdateNetPolRelevant(oldPod, newPod *corev1.Pod) bool {
	return newPod.Status.Phase != oldPod.Status.Phase ||
		newPod.Status.PodIP != oldPod.Status.PodIP ||
		!reflect.DeepEqual(newPod.Status.PodIPs, oldPod.Status.PodIPs) ||
		newPod.Status.HostIP != oldPod.Status.HostIP ||
		!reflect.DeepEqual(newPod.Labels, oldPod.Labels)
}

// GetNodeIP returns the most valid external facing IP address for a node.
// Order of preference:
// 1. NodeInternalIP
// 2. NodeExternalIP (Only set on cloud providers usually)
func GetNodeIP(node *corev1.Node) (net.IP, error) {
	addresses := node.Status.Addresses
	addressMap := make(map[corev1.NodeAddressType][]corev1.NodeAddress)
	for i := range addresses {
		addressMap[addresses[i].Type] = append(addressMap[addresses[i].Type], addresses[i])
	}
	if addresses, ok := addressMap[corev1.NodeInternalIP]; ok {
		return net.ParseIP(addresses[0].Address), nil
	}
	if addresses, ok := addressMap[corev1.NodeExternalIP]; ok {
		return net.ParseIP(addresses[0].Address), nil
	}
	return nil, fmt.Errorf("host IP unknown")
}
