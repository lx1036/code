package defaultpreemption

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"

	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	"k8s-lx1036/k8s/scheduler/pkg/metrics"
	"k8s-lx1036/k8s/scheduler/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	corelisters "k8s.io/client-go/listers/core/v1"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

// INFO: 抢占preemption plugin
//  (1)判断该 pod 是否可以抢占，比如该 pod 抢占策略是不可抢占的，就不抢占
//  (2)找出所有可以被抢占的节点，但最多100台节点，只要改节点上有 Pod 优先级比当前 Pod 优先级低，该节点就是可以被抢占的节点。
//  当然，该 Pod 得经过针对这个 Node 的 Filter 走一遍，比如该 Node 资源还够不够。
//  (3)根据规则找出最优的可被抢占的节点，比如可以被驱逐的 victims 数量最少，优先级总和最小，node 上高优先级 Pod 数量最少等等
//  (4)驱逐 victim pod

const (
	// Name of the plugin used in the plugin registry and configurations.
	Name = "DefaultPreemption"
)

// DefaultPreemption is a PostFilter plugin implements the preemption logic.
type DefaultPreemption struct {
	framework *frameworkruntime.Framework
	args      configv1.DefaultPreemptionArgs
	podLister corelisters.PodLister
}

func New(dpArgs runtime.Object, framework *frameworkruntime.Framework) (framework.Plugin, error) {
	args, ok := dpArgs.(*configv1.DefaultPreemptionArgs)
	if !ok {
		return nil, fmt.Errorf("got args of type %T, want *DefaultPreemptionArgs", dpArgs)
	}
	pl := DefaultPreemption{
		framework: framework,
		args:      *args,
		podLister: framework.SharedInformerFactory().Core().V1().Pods().Lister(),
	}

	return &pl, nil
}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *DefaultPreemption) Name() string {
	return Name
}

// PostFilter INFO: 这里的 nodeStatus 是 pod 经过 PreFilter/Filter plugins 失败的 node 集合，很正常，抢占肯定抢占这些 nodes 上的 pod
func (pl *DefaultPreemption) PostFilter(ctx context.Context, state *framework.CycleState, pod *corev1.Pod,
	nodeStatus framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	defer func() {
		metrics.PreemptionAttempts.Inc()
	}()

	bestCandidate, err := pl.preempt(ctx, state, pod, nodeStatus)
	if err != nil {
		return nil, framework.NewStatus(framework.Error, err.Error())
	}
	if bestCandidate == "" {
		return framework.NewPostFilterResultWithNominatedNode(""), framework.NewStatus(framework.Unschedulable)
	}

	return framework.NewPostFilterResultWithNominatedNode(bestCandidate), framework.NewStatus(framework.Success)
}

func (pl *DefaultPreemption) preempt(ctx context.Context, state *framework.CycleState, pod *corev1.Pod,
	nodeStatus framework.NodeToStatusMap) (string, error) {
	// INFO:(1) Ensure the preemptor is eligible to preempt other pods.
	var err error
	pod, err = pl.podLister.Pods(pod.Namespace).Get(pod.Name)
	if err != nil {
		return "", fmt.Errorf(fmt.Sprintf("Getting the updated preemptor pod %s/%s error", pod.Namespace, pod.Name))
	}
	if !pl.PodEligibleToPreemptOthers(pod, nodeStatus[pod.Status.NominatedNodeName]) {
		return "", fmt.Errorf(fmt.Sprintf("Pod %s/%s is not eligible for more preemption", pod.Namespace, pod.Name))
	}

	// INFO:(2) Find all preemption candidates.
	candidates, _, err := pl.findCandidates(ctx, pod, nodeStatus, state)
	if err != nil || len(candidates) == 0 {
		return "", err
	}

	// INFO: (3) Find the best candidate.
	bestCandidate := pl.SelectCandidate(candidates)
	if bestCandidate == nil || len(bestCandidate.Name()) == 0 {
		return "", fmt.Errorf("no candidate node for preemption")
	}

	// INFO: (4) Perform preparation work before nominating the selected candidate.
	if status := pl.prepareCandidate(bestCandidate, pod, pl.Name()); !status.IsSuccess() {
		return "", fmt.Errorf(fmt.Sprintf("%s", status.Message()))
	}

	return bestCandidate.Name(), nil // success
}

// PodEligibleToPreemptOthers INFO: @see https://kubernetes.io/zh/docs/concepts/configuration/pod-priority-preemption/#non-preempting-priority-class
func (pl *DefaultPreemption) PodEligibleToPreemptOthers(pod *corev1.Pod, nominatedNodeStatus *framework.Status) bool {
	// INFO: 非抢占式的 pod 不需要抢占，返回 false
	if pod.Spec.PreemptionPolicy != nil && *pod.Spec.PreemptionPolicy == corev1.PreemptNever {
		klog.V(5).Infof("Pod %v/%v is not eligible for preemption because it has a preemptionPolicy of %v", pod.Namespace, pod.Name, corev1.PreemptNever)
		return false
	}

	nodeInfos := pl.framework.SnapshotSharedLister().NodeInfos()
	nominatedNodeName := pod.Status.NominatedNodeName // 如果之前已经抢占过了，则检查下
	if len(nominatedNodeName) > 0 {
		// INFO: nominatedNodeName 是不可调度的，可以抢占？
		if nominatedNodeStatus.Code() == framework.UnschedulableAndUnresolvable {
			return true
		}

		// nominatedNodeName 有 terminating pod 了，则不需要抢占
		if nodeInfo, _ := nodeInfos.Get(nominatedNodeName); nodeInfo != nil {
			podPriority := corev1helpers.PodPriority(pod)
			for _, p := range nodeInfo.Pods {
				if p.Pod.DeletionTimestamp != nil && corev1helpers.PodPriority(p.Pod) < podPriority {
					// There is a terminating pod on the nominated node.
					return false
				}
			}
		}
	}

	return true
}

type Candidate struct {
	victims *extenderv1.Victims
	name    string
}

func (s *Candidate) Name() string {
	return s.name
}
func (s *Candidate) Victims() *extenderv1.Victims {
	return s.victims
}

func (pl *DefaultPreemption) findCandidates(ctx context.Context, pod *corev1.Pod,
	nodeStatus framework.NodeToStatusMap, state *framework.CycleState) ([]*Candidate, framework.NodeToStatusMap, error) {
	allNodes, err := pl.framework.SnapshotSharedLister().NodeInfos().List()
	if err != nil {
		return nil, nil, err
	}
	if len(allNodes) == 0 {
		return nil, nil, errors.New("no nodes available")
	}
	potentialNodes, unschedulableAndUnresolvableNodeStatus := nodesWherePreemptionMightHelp(allNodes, nodeStatus)
	if len(potentialNodes) == 0 { // 如果 allNodes 都是 UnschedulableAndUnresolvable
		klog.V(3).InfoS("Preemption will not help schedule pod on any node", "pod", klog.KObj(pod))
		if err := util.ClearNominatedNodeName(pl.framework.ClientSet(), pod); err != nil {
			klog.ErrorS(err, "Cannot clear 'NominatedNodeName' field of pod", "pod", klog.KObj(pod))
			// We do not return as this error is not critical.
		}
		return nil, unschedulableAndUnresolvableNodeStatus, nil
	}

	offset, numCandidates := pl.GetOffsetAndNumCandidates(int32(len(potentialNodes))) // offset=10台, 最小100台
	candidates, nStatus, err := pl.DryRunPreemption(ctx, pod, potentialNodes, offset, numCandidates, state)
	for node, status := range unschedulableAndUnresolvableNodeStatus {
		nStatus[node] = status
	}
	return candidates, nStatus, err
}

func nodesWherePreemptionMightHelp(nodes []*framework.NodeInfo, m framework.NodeToStatusMap) ([]*framework.NodeInfo, framework.NodeToStatusMap) {
	var potentialNodes []*framework.NodeInfo
	nodeStatuses := make(framework.NodeToStatusMap)
	for _, node := range nodes {
		name := node.Node().Name
		if m[name].Code() == framework.UnschedulableAndUnresolvable {
			nodeStatuses[node.Node().Name] = framework.NewStatus(framework.UnschedulableAndUnresolvable, "Preemption is not helpful for scheduling")
			continue
		}
		potentialNodes = append(potentialNodes, node)
	}
	return potentialNodes, nodeStatuses
}

type candidateList struct {
	idx   int32
	items []*Candidate
}

func newCandidateList(size int32) *candidateList {
	return &candidateList{idx: -1, items: make([]*Candidate, size)}
}
func (cl *candidateList) add(c *Candidate) {
	if idx := atomic.AddInt32(&cl.idx, 1); idx < int32(len(cl.items)) {
		cl.items[idx] = c
	}
}
func (cl *candidateList) size() int32 {
	n := atomic.LoadInt32(&cl.idx) + 1
	if n >= int32(len(cl.items)) {
		n = int32(len(cl.items))
	}
	return n
}
func (cl *candidateList) get() []*Candidate {
	return cl.items[:cl.size()]
}

func (pl *DefaultPreemption) DryRunPreemption(ctx context.Context, pod *corev1.Pod, potentialNodes []*framework.NodeInfo,
	offset int32, numCandidates int32, state *framework.CycleState) ([]*Candidate, framework.NodeToStatusMap, error) {
	// INFO: 为了高效率，这里通过多个并发去处理 potentialNodes，而不是一个个去处理 node。
	//  即把 potentialNodes 切成多个 piece，然后每一个 piece 里再一个个去 SelectVictimsOnNode(nodeInfo)
	var statusesLock sync.Mutex
	var errs []error
	nodeStatuses := make(framework.NodeToStatusMap)
	nonViolatingCandidates := newCandidateList(numCandidates) // numCandidates=100
	parallelCtx, cancel := context.WithCancel(ctx)
	checkNode := func(i int) {
		nodeInfoCopy := potentialNodes[(int(offset)+i)%len(potentialNodes)].Clone() // potentialNodes 包含所有 nodes
		stateCopy := state.Clone()
		pods, status := pl.SelectVictimsOnNode(ctx, stateCopy, pod, nodeInfoCopy)
		if status.IsSuccess() && len(pods) != 0 {
			victims := extenderv1.Victims{
				Pods: pods,
			}
			c := &Candidate{
				victims: &victims,
				name:    nodeInfoCopy.Node().Name,
			}
			nonViolatingCandidates.add(c)
			nvcSize := nonViolatingCandidates.size()
			if nvcSize > 0 && nvcSize >= numCandidates { // INFO: 尽管从 450 nodes 中选取，但是满足了 100 个就跳出
				cancel()
			}
			return
		}
		if status.IsSuccess() && len(pods) == 0 {
			status = framework.AsStatus(fmt.Errorf("expected at least one victim pod on node %q", nodeInfoCopy.Node().Name))
		}

		statusesLock.Lock()
		if status.Code() == framework.Error {
			errs = append(errs, status.AsError())
		}
		nodeStatuses[nodeInfoCopy.Node().Name] = status
		statusesLock.Unlock()
	}

	pl.framework.Parallelizer().Until(parallelCtx, len(potentialNodes), checkNode) // block
	return nonViolatingCandidates.get(), nodeStatuses, utilerrors.NewAggregate(errs)
}

func (pl *DefaultPreemption) SelectVictimsOnNode(
	ctx context.Context,
	state *framework.CycleState,
	pod *corev1.Pod,
	nodeInfo *framework.NodeInfo) ([]*corev1.Pod, *framework.Status) {
	removePod := func(rpi *framework.PodInfo) error {
		if err := nodeInfo.RemovePod(rpi.Pod); err != nil {
			return err
		}
		// 注意：还得去掉 PreFilter RemovePod???
		status := pl.framework.RunPreFilterExtensionRemovePod(ctx, state, pod, rpi, nodeInfo)
		if !status.IsSuccess() {
			return status.AsError()
		}
		return nil
	}
	addPod := func(api *framework.PodInfo) error {
		nodeInfo.AddPodInfo(api)
		status := pl.framework.RunPreFilterExtensionAddPod(ctx, state, pod, api, nodeInfo)
		if !status.IsSuccess() {
			return status.AsError()
		}
		return nil
	}

	// INFO: 从当前 nodeInfo 中找出低优的 pod
	var potentialVictims []*framework.PodInfo
	podPriority := corev1helpers.PodPriority(pod)
	for _, pi := range nodeInfo.Pods {
		if corev1helpers.PodPriority(pi.Pod) < podPriority {
			potentialVictims = append(potentialVictims, pi)
			if err := removePod(pi); err != nil { // 从 nodeInfo 中去掉 lowPriority Pod
				return nil, framework.AsStatus(err)
			}
		}
	}
	if len(potentialVictims) == 0 {
		message := fmt.Sprintf("No preemption victims found for incoming pod")
		return nil, framework.NewStatus(framework.UnschedulableAndUnresolvable, message)
	}

	// INFO: 再次重新检查 pod，尽可能不把当前 pod 作为 victimPod
	var victims []*corev1.Pod
	reprievePod := func(pi *framework.PodInfo) (bool, error) {
		if err := addPod(pi); err != nil {
			return false, err
		}
		// INFO: 这里把这台 nodeInfo 上的优先级更高的 pods 剔除出去
		status := pl.framework.RunFilterPluginsWithNominatedPods(ctx, state, pod, nodeInfo)
		success := status.IsSuccess()
		if !success {
			if err := removePod(pi); err != nil {
				return false, err
			}
			rpi := pi.Pod
			victims = append(victims, rpi)
			klog.V(5).InfoS("Pod is a potential preemption victim on node", "pod", klog.KObj(rpi), "node", klog.KObj(nodeInfo.Node()))
		}
		return success, nil
	}
	for _, p := range potentialVictims {
		if _, err := reprievePod(p); err != nil {
			return nil, framework.AsStatus(err)
		}
	}

	return victims, framework.NewStatus(framework.Success)
}

// SelectCandidate INFO: 优选 nodes，pod优先级之和最低，pod数量最少的，pod最旧的
func (pl *DefaultPreemption) SelectCandidate(candidates []*Candidate) *Candidate {
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	victimsMap := pl.CandidatesToVictimsMap(candidates)
	candidateNode := pickOneNodeForPreemption(victimsMap)
	if victims := victimsMap[candidateNode]; victims != nil {
		return &Candidate{
			victims: victims,
			name:    candidateNode,
		}
	}

	// We shouldn't reach here.
	klog.ErrorS(errors.New("no candidate selected"), "Should not reach here", "candidates", candidates)
	// To not break the whole flow, return the first candidate.
	return candidates[0]
}

// CandidatesToVictimsMap INFO: 优选 nodes，pod优先级之和最低，pod数量最少的，pod最旧的
func (pl *DefaultPreemption) CandidatesToVictimsMap(candidates []*Candidate) map[string]*extenderv1.Victims {
	m := make(map[string]*extenderv1.Victims)
	for _, c := range candidates {
		m[c.Name()] = c.Victims()
	}
	return m
}

// INFO: 挑选规则：
//  1. A node with minimum highest priority victim is picked.
//  2. Ties are broken by sum of priorities of all victims.
//  3. If there are still ties, node with the minimum number of victims is picked.
//  4. If there are still ties, node with the latest start time of all highest priority victims is picked.
//  5. If there are still ties, the first such node is picked (sort of randomly).
func pickOneNodeForPreemption(nodesToVictims map[string]*extenderv1.Victims) string {
	if len(nodesToVictims) == 0 {
		return ""
	}
	minNumPDBViolatingPods := int64(math.MaxInt32)
	var minNodes1 []string
	lenNodes1 := 0
	for node, victims := range nodesToVictims {
		numPDBViolatingPods := victims.NumPDBViolations
		if numPDBViolatingPods < minNumPDBViolatingPods {
			minNumPDBViolatingPods = numPDBViolatingPods
			minNodes1 = nil
			lenNodes1 = 0
		}
		if numPDBViolatingPods == minNumPDBViolatingPods {
			minNodes1 = append(minNodes1, node)
			lenNodes1++
		}
	}
	if lenNodes1 == 1 {
		return minNodes1[0]
	}

	// 找出 pod 优先级最低的那个 node
	minHighestPriority := int32(math.MaxInt32)
	var minNodes2 = make([]string, lenNodes1)
	lenNodes2 := 0
	for i := 0; i < lenNodes1; i++ {
		node := minNodes1[i]
		victims := nodesToVictims[node]
		// highestPodPriority is the highest priority among the victims on this node.
		highestPodPriority := corev1helpers.PodPriority(victims.Pods[0])
		if highestPodPriority < minHighestPriority {
			minHighestPriority = highestPodPriority
			lenNodes2 = 0
		}
		if highestPodPriority == minHighestPriority {
			minNodes2[lenNodes2] = node
			lenNodes2++
		}
	}
	if lenNodes2 == 1 {
		return minNodes2[0]
	}

	// 如果还没找到，就找优先级之和最小的那个 node
	minSumPriorities := int64(math.MaxInt64)
	lenNodes1 = 0
	for i := 0; i < lenNodes2; i++ {
		var sumPriorities int64
		node := minNodes2[i]
		for _, pod := range nodesToVictims[node].Pods {
			// We add MaxInt32+1 to all priorities to make all of them >= 0. This is
			// needed so that a node with a few pods with negative priority is not
			// picked over a node with a smaller number of pods with the same negative
			// priority (and similar scenarios).
			sumPriorities += int64(corev1helpers.PodPriority(pod)) + int64(math.MaxInt32+1)
		}
		if sumPriorities < minSumPriorities {
			minSumPriorities = sumPriorities
			lenNodes1 = 0
		}
		if sumPriorities == minSumPriorities {
			minNodes1[lenNodes1] = node
			lenNodes1++
		}
	}
	if lenNodes1 == 1 {
		return minNodes1[0]
	}

	// pod 数量最少的那个 node
	minNumPods := math.MaxInt32
	lenNodes2 = 0
	for i := 0; i < lenNodes1; i++ {
		node := minNodes1[i]
		numPods := len(nodesToVictims[node].Pods)
		if numPods < minNumPods {
			minNumPods = numPods
			lenNodes2 = 0
		}
		if numPods == minNumPods {
			minNodes2[lenNodes2] = node
			lenNodes2++
		}
	}
	if lenNodes2 == 1 {
		return minNodes2[0]
	}

	// 最旧的那些 pod 所在的 node
	latestStartTime := util.GetEarliestPodStartTime(nodesToVictims[minNodes2[0]])
	if latestStartTime == nil {
		// If the earliest start time of all pods on the 1st node is nil, just return it,
		// which is not expected to happen.
		klog.ErrorS(errors.New("earliestStartTime is nil for node"), "Should not reach here", "node", klog.KRef("", minNodes2[0]))
		return minNodes2[0]
	}
	nodeToReturn := minNodes2[0]
	for i := 1; i < lenNodes2; i++ {
		node := minNodes2[i]
		// Get earliest start time of all pods on the current node.
		earliestStartTimeOnNode := util.GetEarliestPodStartTime(nodesToVictims[node])
		if earliestStartTimeOnNode == nil {
			klog.ErrorS(errors.New("earliestStartTime is nil for node"), "Should not reach here", "node", klog.KRef("", node))
			continue
		}
		if earliestStartTimeOnNode.After(latestStartTime.Time) {
			latestStartTime = earliestStartTimeOnNode
			nodeToReturn = node
		}
	}

	return nodeToReturn
}

// prepareCandidate does some preparation work before nominating the selected candidate:
// - Evict the victim pods
// - Reject the victim pods if they are in waitingPod map
// - Clear the low-priority pods' nominatedNodeName status if needed
func (pl *DefaultPreemption) prepareCandidate(c *Candidate, pod *corev1.Pod, pluginName string) *framework.Status {
	for _, victim := range c.Victims().Pods {
		// If the victim is a WaitingPod, send a reject message to the PermitPlugin.
		// Otherwise we should delete the victim.
		if waitingPod := pl.framework.GetWaitingPod(victim.UID); waitingPod != nil {
			waitingPod.Reject(pluginName, "preempted")
		} else if err := util.DeletePod(pl.framework.ClientSet(), victim); err != nil {
			klog.ErrorS(err, "Preempting pod", "pod", klog.KObj(victim), "preemptor", klog.KObj(pod))
			return framework.AsStatus(err)
		}
		pl.framework.EventRecorder().Eventf(victim, pod, corev1.EventTypeNormal, "Preempted", "Preempting", "Preempted by %v/%v on node %v",
			pod.Namespace, pod.Name, c.Name())
	}

	nominatedPods := pl.getLowerPriorityNominatedPods(pod, c.Name())
	if err := util.ClearNominatedNodeName(pl.framework.ClientSet(), nominatedPods...); err != nil {
		klog.ErrorS(err, "Cannot clear 'NominatedNodeName' field")
		// We do not return as this error is not critical.
	}

	return nil
}

// 优先级比 pod 小的 NominatedPods
func (pl *DefaultPreemption) getLowerPriorityNominatedPods(pod *corev1.Pod, nodeName string) []*corev1.Pod {
	podInfos := pl.framework.PodNominator.NominatedPodsForNode(nodeName)
	if len(podInfos) == 0 {
		return nil
	}

	var lowerPriorityPods []*corev1.Pod
	podPriority := corev1helpers.PodPriority(pod)
	for _, pi := range podInfos {
		if corev1helpers.PodPriority(pi.Pod) < podPriority {
			lowerPriorityPods = append(lowerPriorityPods, pi.Pod)
		}
	}
	return lowerPriorityPods
}

func (pl *DefaultPreemption) GetOffsetAndNumCandidates(numNodes int32) (int32, int32) {
	return rand.Int31n(numNodes), pl.calculateNumCandidates(numNodes)
}
func (pl *DefaultPreemption) calculateNumCandidates(numNodes int32) int32 {
	n := (numNodes * pl.args.MinCandidateNodesPercentage) / 100 // 450 * 10/100
	if n < pl.args.MinCandidateNodesAbsolute {                  // 最小 100 台
		n = pl.args.MinCandidateNodesAbsolute
	}
	if n > numNodes {
		n = numNodes
	}
	return n
}
