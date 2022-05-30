package controller

import (
	"bytes"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strconv"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	utilproxy "k8s.io/kubernetes/pkg/proxy/util"
	"k8s.io/kubernetes/pkg/util/iptables"
	"k8s.io/utils/exec"
)

// INFO: @see https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/iptables/proxier.go

const (
	kubeServicesChain         iptables.Chain = "KUBE-SERVICES"
	kubeExternalServicesChain iptables.Chain = "KUBE-EXTERNAL-SERVICES"
	kubeNodePortsChain        iptables.Chain = "KUBE-NODEPORTS"
	kubeForwardChain          iptables.Chain = "KUBE-FORWARD"
	kubePostroutingChain      iptables.Chain = "KUBE-POSTROUTING"

	KubeMarkDropChain iptables.Chain = "KUBE-MARK-DROP"

	// KubeMarkMasqChain IP地址伪装，endpoint 做地址伪装
	KubeMarkMasqChain iptables.Chain = "KUBE-MARK-MASQ"
)

type iptablesJumpChain struct {
	table     iptables.Table
	dstChain  iptables.Chain
	srcChain  iptables.Chain
	comment   string
	extraArgs []string
}

var iptablesJumpChains = []iptablesJumpChain{
	{iptables.TableFilter, kubeExternalServicesChain, iptables.ChainInput, "kubernetes externally-visible service portals", []string{"-m", "conntrack", "--ctstate", "NEW"}},
	{iptables.TableFilter, kubeExternalServicesChain, iptables.ChainForward, "kubernetes externally-visible service portals", []string{"-m", "conntrack", "--ctstate", "NEW"}},
	{iptables.TableFilter, kubeNodePortsChain, iptables.ChainInput, "kubernetes health check service ports", nil},
	{iptables.TableFilter, kubeServicesChain, iptables.ChainForward, "kubernetes service portals", []string{"-m", "conntrack", "--ctstate", "NEW"}},
	{iptables.TableFilter, kubeServicesChain, iptables.ChainOutput, "kubernetes service portals", []string{"-m", "conntrack", "--ctstate", "NEW"}},
	{iptables.TableFilter, kubeForwardChain, iptables.ChainForward, "kubernetes forwarding rules", nil},

	{iptables.TableNAT, kubeServicesChain, iptables.ChainOutput, "kubernetes service portals", nil},
	{iptables.TableNAT, kubeServicesChain, iptables.ChainPrerouting, "kubernetes service portals", nil},
	{iptables.TableNAT, kubePostroutingChain, iptables.ChainPostrouting, "kubernetes postrouting rules", nil},
}
var iptablesEnsureChains = []struct {
	table iptables.Table
	chain iptables.Chain
}{
	{iptables.TableNAT, KubeMarkDropChain},
}

type NetworkServiceController struct {
	sync.Mutex

	iptablesCmdHandler iptables.Interface
	svcLister          cache.Indexer
	epLister           cache.Indexer

	existingFilterChainsData *bytes.Buffer
	iptablesData             *bytes.Buffer
	filterChains             utilproxy.LineBuffer
	filterRules              utilproxy.LineBuffer
	natChains                utilproxy.LineBuffer
	natRules                 utilproxy.LineBuffer

	masqueradeMark string
	masqueradeAll  bool

	syncChan chan struct{}
}

func NewNetworkPolicyController(
	clientset kubernetes.Interface,
	svcInformer cache.SharedIndexInformer,
	epInformer cache.SharedIndexInformer,
	masqueradeBit int, // default 14
	masqueradeAll bool, // "If using the pure iptables proxy, SNAT all traffic sent via Service cluster IPs (this not commonly needed)"
) (*NetworkServiceController, error) {

	// Generate the masquerade mark to use for SNAT rules.
	// If using the pure iptables proxy, the bit of the fwmark space to mark packets requiring SNAT with.  Must be within the range [0, 31].
	masqueradeValue := 1 << uint(masqueradeBit)
	masqueradeMark := fmt.Sprintf("%#08x", masqueradeValue)
	c := &NetworkServiceController{
		svcLister:          svcInformer.GetIndexer(),
		epLister:           epInformer.GetIndexer(),
		iptablesCmdHandler: iptables.New(exec.New(), iptables.ProtocolIPv4),

		existingFilterChainsData: bytes.NewBuffer(nil),
		iptablesData:             bytes.NewBuffer(nil),

		masqueradeMark: masqueradeMark,
		masqueradeAll:  masqueradeAll,

		syncChan: make(chan struct{}, 2), // buffer chan，因为 service 互不影响，channel item 可以多个, @see NetworkPolicyController
	}

	svcInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			service, ok := obj.(*corev1.Service)
			if !ok {
				return false
			}
			if IsHeadlessService(service) || IsExternalNameService(service) { // skip headless service
				return false
			}
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				service, ok := obj.(*corev1.Service)
				if !ok {
					return
				}
				c.onServiceUpdate(service)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				service, ok := newObj.(*corev1.Service)
				if !ok {
					return
				}
				c.onServiceUpdate(service)
			},
			DeleteFunc: func(obj interface{}) {
				service, ok := obj.(*corev1.Service)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						klog.Errorf("unexpected object type: %v", obj)
						return
					}
					if service, ok = tombstone.Obj.(*corev1.Service); !ok {
						klog.Errorf("unexpected object type: %v", obj)
						return
					}
				}

				c.onServiceUpdate(service)
			},
		},
	})

}

func (controller *NetworkServiceController) Run(stopCh <-chan struct{}) {
	t := time.NewTicker(controller.syncPeriod)
	defer t.Stop()

	select {
	case <-stopCh:
		klog.Info("Shutting down network services controller")
		return
	default:
		err := controller.doSync()
		if err != nil {
			klog.Fatalf(fmt.Sprintf("Failed to perform initial full sync %v", err))
		}
		controller.readyForUpdates = true
	}

	for {
		select {
		case <-stopCh:
			controller.readyForUpdates = false
			return

		case <-controller.syncChan:
			// We call the component pieces of doSync() here because for methods that send this on the channel they
			// have already done expensive pieces of the doSync() method like building service and endpoint info
			// and we don't want to duplicate the effort, so this is a slimmer version of doSync()
			controller.Lock()
			controller.syncService(controller.serviceMap, controller.endpointMap)
			controller.syncHairpinIptablesRules()

			controller.Unlock()

		case <-t.C:
			controller.doSync()
		}
	}
}

func (controller *NetworkServiceController) sync() {
	select {
	case controller.syncChan <- struct{}{}:
	default:
		klog.Infof("Already pending sync, dropping request")
	}
}

func (controller *NetworkServiceController) syncService(serviceMap serviceInfoMap, endpointMap endpointInfoMap) {

	// (1) setup jump chains
	// INFO: filter 表主要在：input/output/forward chain，还有自定义的: KUBE-SERVICES/KUBE-EXTERNAL-SERVICES/KUBE-FORWARD/KUBE-NODEPORTS
	//  forward chain -> KUBE-SERVICES/KUBE-EXTERNAL-SERVICES/KUBE-FORWARD, output chain -> KUBE-SERVICES, input chain -> KUBE-EXTERNAL-SERVICES/KUBE-NODEPORTS
	//  nat 表主要在: prerouting/postrouting/output, 还有自定义的 KUBE-SERVICES/KUBE-POSTROUTING/KUBE-MARK-DROP
	//  prerouting -> KUBE-SERVICES, postrouting -> KUBE-POSTROUTING, output -> KUBE-SERVICES
	for _, jump := range iptablesJumpChains {
		if _, err := controller.iptablesCmdHandler.EnsureChain(jump.table, jump.dstChain); err != nil {
			klog.ErrorS(err, "Failed to ensure chain exists", "table", jump.table, "chain", jump.dstChain)
			return
		}
		args := append(jump.extraArgs, "-m", "comment", "--comment", jump.comment, "-j", string(jump.dstChain))
		if _, err := controller.iptablesCmdHandler.EnsureRule(iptables.Prepend, jump.table, jump.srcChain, args...); err != nil {
			klog.ErrorS(err, "Failed to ensure chain jumps", "table", jump.table, "srcChain", jump.srcChain, "dstChain", jump.dstChain)
			return
		}
	}
	// ensure KUBE-MARK-DROP chain exist but do not change any rules
	for _, ch := range iptablesEnsureChains {
		if _, err := controller.iptablesCmdHandler.EnsureChain(ch.table, ch.chain); err != nil {
			klog.ErrorS(err, "Failed to ensure chain exists", "table", ch.table, "chain", ch.chain)
			return
		}
	}

	// 综上，filter 表内包含多个 chain 的 rule
	existingFilterChains := make(map[iptables.Chain][]byte)
	controller.existingFilterChainsData.Reset()
	err := controller.iptablesCmdHandler.SaveInto(iptables.TableFilter, controller.existingFilterChainsData)
	if err != nil {
		klog.Errorf(fmt.Sprintf("Failed to execute iptables-save, syncing all rules"))
	} else {
		existingFilterChains = iptables.GetChainLines(iptables.TableFilter, controller.existingFilterChainsData.Bytes())
	}
	existingNatChains := make(map[iptables.Chain][]byte)
	controller.iptablesData.Reset()
	err = controller.iptablesCmdHandler.SaveInto(iptables.TableNAT, controller.iptablesData)
	if err != nil {
		klog.Errorf(fmt.Sprintf("Failed to execute iptables-save, syncing all rules"))
	} else {
		existingNatChains = iptables.GetChainLines(iptables.TableNAT, controller.iptablesData.Bytes())
	}
	// Reset all buffers used later.
	// This is to avoid memory reallocations and thus improve performance.
	controller.filterChains.Reset()
	controller.filterRules.Reset()
	controller.natChains.Reset()
	controller.natRules.Reset()
	// 综上已经创建了这些自定义 chain
	for _, chainName := range []iptables.Chain{kubeServicesChain, kubeExternalServicesChain, kubeForwardChain, kubeNodePortsChain} {
		if chain, ok := existingFilterChains[chainName]; ok {
			controller.filterChains.WriteBytes(chain) // 已经创建了，这里只需要写 "KUBE-SERVICES\n"
		} else {
			controller.filterChains.Write(iptables.MakeChainLine(chainName)) // 新建则需要写 ":KUBE-SERVICES - [0:0]\n"
		}
	}
	for _, chainName := range []iptables.Chain{kubeServicesChain, kubeNodePortsChain, kubePostroutingChain, KubeMarkMasqChain} {
		if chain, ok := existingNatChains[chainName]; ok {
			controller.natChains.WriteBytes(chain)
		} else {
			controller.natChains.Write(iptables.MakeChainLine(chainName))
		}
	}

	// Install the kubernetes-specific postrouting rules. We use a whole chain for
	// this so that it is easier to flush and change, for example if the mark
	// value should ever change.
	controller.natRules.Write(
		"-A", string(kubePostroutingChain),
		"-m", "mark", "!", "--mark", fmt.Sprintf("%s/%s", controller.masqueradeMark, controller.masqueradeMark),
		"-j", "RETURN",
	)
	// Clear the mark to avoid re-masquerading if the packet re-traverses the network stack.
	controller.natRules.Write(
		"-A", string(kubePostroutingChain),
		"-j", "MARK", "--xor-mark", controller.masqueradeMark,
	)
	masqRule := []string{
		"-A", string(kubePostroutingChain),
		"-m", "comment", "--comment", `"kubernetes service traffic requiring SNAT"`,
		"-j", "MASQUERADE",
	}
	if controller.iptablesCmdHandler.HasRandomFully() {
		masqRule = append(masqRule, "--random-fully")
	}
	controller.natRules.Write(masqRule)
	// Install the kubernetes-specific masquerade mark rule. We use a whole chain for
	// this so that it is easier to flush and change, for example if the mark
	// value should ever change.
	controller.natRules.Write(
		"-A", string(KubeMarkMasqChain),
		"-j", "MARK", "--or-mark", controller.masqueradeMark,
	)

	// Build iptables rules for each service-port.
	for svcName, svc := range controller.serviceInfoMap {
		allEndpoints := controller.endpointsMap[svcName]

		// Generate the per-endpoint chains.
		for _, ep := range allLocallyReachableEndpoints {

			args = append(args[:0], "-A", string(endpointChain))

			// Handle traffic that loops back to the originator with SNAT.
			controller.natRules.Write(args, "-s", epInfo.IP(), "-j", string(KubeMarkMasqChain))
			// Update client-affinity lists.
			if svcInfo.SessionAffinityType() == corev1.ServiceAffinityClientIP {
				args = append(args, "-m", "recent", "--name", string(endpointChain), "--set")
			}
			// DNAT to final destination.
			args = append(args, "-m", protocol, "-p", protocol, "-j", "DNAT", "--to-destination", epInfo.Endpoint)
			controller.natRules.Write(args)
		}

		// Capture the clusterIP
		if hasEndpoints {

		} else {
			// No endpoints.
			controller.filterRules.Write(
				"-A", string(kubeServicesChain),
				"-m", "comment", "--comment", fmt.Sprintf(`"%s has no endpoints"`, svcNameString),
				"-m", protocol, "-p", protocol,
				"-d", svcInfo.ClusterIP().String(),
				"--dport", strconv.Itoa(svcInfo.Port()),
				"-j", "REJECT",
			)
		}

		if svcInfo.UsesClusterEndpoints() {
			// Write rules jumping from clusterPolicyChain to clusterEndpoints
			controller.writeServiceToEndpointRules(svcNameString, svcInfo, clusterPolicyChain, clusterEndpoints, args)
		}
		if svcInfo.UsesLocalEndpoints() {
			if len(localEndpoints) != 0 {
				// Write rules jumping from localPolicyChain to localEndpointChains
				controller.writeServiceToEndpointRules(svcNameString, svcInfo, localPolicyChain, localEndpoints, args)
			} else {

			}
		}
	}

	// Finally, tail-call to the nodeports chain.  This needs to be after all other service portal rules.
	for address := range nodeAddresses {
		// create nodeport rules for each IP one by one
		controller.natRules.Write(
			"-A", string(kubeServicesChain),
			"-m", "comment", "--comment", `"kubernetes service nodeports; NOTE: this must be the last rule in this chain"`,
			"-d", address,
			"-j", string(kubeNodePortsChain))
	}

	// Drop the packets in INVALID state, which would potentially cause
	// unexpected connection reset.
	// https://github.com/kubernetes/kubernetes/issues/74839
	controller.filterRules.Write(
		"-A", string(kubeForwardChain),
		"-m", "conntrack",
		"--ctstate", "INVALID",
		"-j", "DROP",
	)
	// If the masqueradeMark has been added then we want to forward that same
	// traffic, this allows NodePort traffic to be forwarded even if the default
	// FORWARD policy is not accept.
	controller.filterRules.Write(
		"-A", string(kubeForwardChain),
		"-m", "comment", "--comment", `"kubernetes forwarding rules"`,
		"-m", "mark", "--mark", fmt.Sprintf("%s/%s", controller.masqueradeMark, controller.masqueradeMark),
		"-j", "ACCEPT",
	)
	// The following rule ensures the traffic after the initial packet accepted
	// by the "kubernetes forwarding rules" rule above will be accepted.
	controller.filterRules.Write(
		"-A", string(kubeForwardChain),
		"-m", "comment", "--comment", `"kubernetes forwarding conntrack rule"`,
		"-m", "conntrack",
		"--ctstate", "RELATED,ESTABLISHED",
		"-j", "ACCEPT",
	)

	// Sync rules.
	controller.iptablesData.Reset()
	controller.iptablesData.WriteString("*filter\n")
	controller.iptablesData.Write(controller.filterChains.Bytes())
	controller.iptablesData.Write(controller.filterRules.Bytes())
	controller.iptablesData.WriteString("COMMIT\n")

	controller.iptablesData.WriteString("*nat\n")
	controller.iptablesData.Write(controller.natChains.Bytes())
	controller.iptablesData.Write(controller.natRules.Bytes())
	controller.iptablesData.WriteString("COMMIT\n")

	// NOTE: NoFlushTables is used so we don't flush non-kubernetes chains in the table
	// 只会 flush 自定义的那几个 chains
	err = controller.iptablesCmdHandler.RestoreAll(controller.iptablesData.Bytes(), iptables.NoFlushTables, iptables.RestoreCounters)
	if err != nil {

		return
	}

}

func (controller *NetworkServiceController) writeServiceToEndpointRules() {
	// INFO: service session affinity, 使用 iptables recent 模块 @see https://linux.die.net/man/8/iptables
	//  -m recent --name epChainName --rcheck --seconds 60
	//  /proc/net/ipt_recent/* are the current lists of addresses and information about each entry of each list
	if svcInfo.SessionAffinityType() == corev1.ServiceAffinityClientIP {
		for _, ep := range endpoints {
			epInfo, ok := ep.(*endpointsInfo)
			if !ok {
				continue
			}
			comment := fmt.Sprintf(`"%s -> %s"`, svcNameString, epInfo.Endpoint)

			args = append(args[:0],
				"-A", string(svcChain),
			)
			args = controller.appendServiceCommentLocked(args, comment)
			args = append(args, "-m", "recent", "--name", string(epInfo.ChainName),
				"--rcheck", "--seconds", strconv.Itoa(svcInfo.StickyMaxAgeSeconds()), "--reap",
				"-j", string(epInfo.ChainName),
			)
			controller.natRules.Write(args)
		}
	}

	numEndpoints := len(endpoints)
	for i, ep := range endpoints {
		epInfo, ok := ep.(*endpointsInfo)
		if !ok {
			continue
		}
		comment := fmt.Sprintf(`"%s -> %s"`, svcNameString, epInfo.Endpoint)

		args = append(args[:0], "-A", string(svcChain))
		args = controller.appendServiceCommentLocked(args, comment)
		if i < (numEndpoints - 1) {
			// Each rule is a probabilistic match.
			args = append(args,
				"-m", "statistic",
				"--mode", "random",
				"--probability", controller.probability(numEndpoints-i))
		}
		// The final (or only if n == 1) rule is a guaranteed match.
		controller.natRules.Write(args, "-j", string(epInfo.ChainName))
	}
}
