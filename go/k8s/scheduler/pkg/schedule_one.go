package pkg

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"

	"k8s-lx1036/k8s/scheduler/pkg/framework"
	"k8s-lx1036/k8s/scheduler/pkg/framework/parallelize"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	internalqueue "k8s-lx1036/k8s/scheduler/pkg/internal/queue"
	"k8s-lx1036/k8s/scheduler/pkg/util"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

const (
	SchedulerError = "SchedulerError"
)

var (
	ErrNoNodesAvailable = fmt.Errorf("no nodes available to schedule pods")
	clearNominatedNode  = &framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedNodeName: ""}
)

// ScheduleResult represents the result of scheduling a pod.
type ScheduleResult struct {
	// Name of the selected node.
	SuggestedHost string
	// The number of nodes the scheduler evaluated the pod against in the filtering
	// phase and beyond.
	EvaluatedNodes int
	// The number of nodes out of the evaluated ones that fit the pod.
	FeasibleNodes int
}

func (scheduler *Scheduler) scheduleOne(ctx context.Context) {
	podInfo := scheduler.NextPod()
	if podInfo == nil || podInfo.Pod == nil {
		return
	}
	pod := podInfo.Pod
	fwk, err := scheduler.frameworkForPod(pod)
	if err != nil {
		klog.Error(err)
		return // skip to schedule this pod
	}
	if scheduler.skipPodSchedule(fwk, pod) {
		return
	}

	klog.Infof("Attempting to schedule pod: %v/%v", pod.Namespace, pod.Name)

	// INFO: (1) PreFilter/Filter
	//  (2) PreScore/Score
	state := framework.NewCycleState()
	state.SetRecordPluginMetrics(rand.Intn(100) < pluginMetricsSamplePercent) // INFO: 这里逻辑只有10%概率记录 plugin metrics
	podsToActivate := framework.NewPodsToActivate()
	state.Write(framework.PodsToActivateKey, podsToActivate)
	schedulingCycleCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	scheduleResult, err := scheduler.SchedulePod(schedulingCycleCtx, fwk, state, pod) // for test case
	if err != nil {
		// INFO: (3) 如果pod调度失败，则调用 PostFilter plugin 进行抢占
		var nominatingInfo *framework.NominatingInfo
		if fitError, ok := err.(*framework.FitError); ok {
			if !fwk.HasPostFilterPlugins() {
				klog.V(3).Infof("No PostFilter plugins are registered, so no preemption will be performed.")
			} else {
				// INFO: PostFilter plugin 其实就是 defaultpreemption.Name plugin，运行 preemption plugin
				result, status := fwk.RunPostFilterPlugins(ctx, state, pod, fitError.Diagnosis.NodeToStatusMap)
				if status.Code() == framework.Error {
					klog.Errorf("Status after running PostFilter plugins for pod %v/%v: %v", pod.Namespace, pod.Name, status)
				} else {
					fitError.Diagnosis.PostFilterMsg = status.Message()
					klog.V(5).Infof("Status after running PostFilter plugins for pod %v/%v: %v", pod.Namespace, pod.Name, status)
				}
				if status.IsSuccess() && result != nil {
					// INFO: 如果抢占成功，则去更新 pod.Status.NominatedNodeName，但是这次调度周期不会立刻更新 pod.Spec.nodeName，
					// 等待下次调度周期去调度。同时，下次调度周期时 pod.Spec.nodeName 未必就是 pod.Status.NominatedNodeName 这个 node
					// 可以去看 k8s.io/api/core/v1/types.go::NominatedNodeName 字段定义描述
					nominatingInfo = result.NominatingInfo
				}
			}
			// metrics
		} else if err == ErrNoNodesAvailable {
			nominatingInfo = clearNominatedNode
		} else {
			nominatingInfo = clearNominatedNode
			klog.ErrorS(err, "Error selecting node for pod", "pod", klog.KObj(pod))
			//metrics.PodScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
		}

		// INFO: 更新 pod.Status.NominatedNodeName，以及更新 pod.Status.Conditions 便于展示信息
		scheduler.handleSchedulingFailure(fwk, podInfo, err, corev1.PodReasonUnschedulable, nominatingInfo)
		return
	}

	// INFO: 已经经过预选和优选，在下一步进入 bind(异步的) 之前，预先假定认为 assume 该 pod 已经 bind 了，防止出现问题，这么做的原因是 bind 是异步的
	//  设置 pod.Spec.NodeName = scheduleResult.SuggestedHost
	assumedPodInfo := podInfo.DeepCopy()
	assumedPod := assumedPodInfo.Pod
	err = scheduler.assume(assumedPod, scheduleResult.SuggestedHost)
	if err != nil {
		//metrics.PodScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
		scheduler.handleSchedulingFailure(fwk, assumedPodInfo, err, SchedulerError, clearNominatedNode)
		return
	}

	// INFO: (4) Reserve(目前只有 VolumeBinding plugin)
	if sts := fwk.RunReservePluginsReserve(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost); !sts.IsSuccess() {
		//metrics.PodScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
		fwk.RunReservePluginsUnreserve(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if forgetErr := scheduler.SchedulerCache.ForgetPod(assumedPod); forgetErr != nil {
			klog.ErrorS(forgetErr, "Scheduler cache ForgetPod failed")
		}
		scheduler.handleSchedulingFailure(fwk, assumedPodInfo, sts.AsError(), SchedulerError, clearNominatedNode)
		return
	}

	// INFO: (5) Permit(目前还没有对应的 plugin)
	runPermitStatus := fwk.RunPermitPlugins(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
	if runPermitStatus.Code() != framework.Wait && !runPermitStatus.IsSuccess() {
		var reason string
		if runPermitStatus.IsUnschedulable() {
			//metrics.PodUnschedulable(fwk.ProfileName(), metrics.SinceInSeconds(start))
			reason = corev1.PodReasonUnschedulable
		} else {
			//metrics.PodScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
			reason = SchedulerError
		}
		fwk.RunReservePluginsUnreserve(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if forgetErr := scheduler.SchedulerCache.ForgetPod(assumedPod); forgetErr != nil {
			klog.ErrorS(forgetErr, "Scheduler cache ForgetPod failed")
		}
		scheduler.handleSchedulingFailure(fwk, assumedPodInfo, runPermitStatus.AsError(), reason, clearNominatedNode)
		return
	}

	// At the end of a successful scheduling cycle, move pods from unschedulePods/backoffQ to activeQ
	if len(podsToActivate.Map) != 0 {
		scheduler.PriorityQueue.Activate(podsToActivate.Map)
		// Clear the entries after activation.
		podsToActivate.Map = make(map[string]*corev1.Pod)
	}

	// INFO: (6) PreBind(有 VolumeBinding)/Bind(只有 DefaultBind)/PostBind(暂无)
	go func() {
		bindingCycleCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		// WaitOnPermit() 参考 PodGroup plugin，如果一组 pod 内，有一个 pod 调度失败了，
		// 会在 PostFilter 里逐个把其他 pod 给 reject 掉，因为其他 pod 这时已经在 WaitOnPermit()
		waitOnPermitStatus := fwk.WaitOnPermit(bindingCycleCtx, assumedPod)
		if !waitOnPermitStatus.IsSuccess() { // PodGroup plugin PostFilter 会
			var reason string
			if waitOnPermitStatus.IsUnschedulable() {
				reason = corev1.PodReasonUnschedulable
			} else {
				reason = SchedulerError
			}
			// trigger un-reserve plugins to clean up state associated with the reserved Pod
			fwk.RunReservePluginsUnreserve(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
			if forgetErr := scheduler.SchedulerCache.ForgetPod(assumedPod); forgetErr != nil {
				klog.ErrorS(forgetErr, "scheduler cache ForgetPod failed")
			} else {
				defer scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.AssignedPodDelete, func(pod *corev1.Pod) bool {
					return assumedPod.UID != pod.UID
				})
			}
			scheduler.handleSchedulingFailure(fwk, assumedPodInfo, waitOnPermitStatus.AsError(), reason, clearNominatedNode)
			return
		}
		preBindStatus := fwk.RunPreBindPlugins(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if !preBindStatus.IsSuccess() {
			fwk.RunReservePluginsUnreserve(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
			if forgetErr := scheduler.SchedulerCache.ForgetPod(assumedPod); forgetErr != nil {
				klog.ErrorS(forgetErr, "scheduler cache ForgetPod failed")
			} else {
				scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.AssignedPodDelete, nil)
			}
			scheduler.handleSchedulingFailure(fwk, assumedPodInfo, preBindStatus.AsError(), SchedulerError, clearNominatedNode)
			return
		}

		bindStatus := fwk.RunBindPlugins(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if !bindStatus.IsSuccess() {
			fwk.RunReservePluginsUnreserve(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
			if err := scheduler.SchedulerCache.ForgetPod(assumedPod); err != nil {
				klog.ErrorS(err, "scheduler cache ForgetPod failed")
			} else {
				scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.AssignedPodDelete, nil)
			}
			scheduler.handleSchedulingFailure(fwk, assumedPodInfo, fmt.Errorf("binding rejected: %w", err), SchedulerError, clearNominatedNode)
			return
		}

		fwk.RunPostBindPlugins(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		// At the end of a successful binding cycle, move up Pods if needed.
		if len(podsToActivate.Map) != 0 {
			scheduler.PriorityQueue.Activate(podsToActivate.Map)
		}
	}()
}

// skipPodSchedule returns true if we could skip scheduling the pod for specified cases.
func (scheduler *Scheduler) skipPodSchedule(prof *frameworkruntime.Framework, pod *corev1.Pod) bool {
	// Case 1: pod is being deleted.
	if pod.DeletionTimestamp != nil {
		return true
	}

	// Case 2: pod that has been assumed could be skipped.
	isAssumed, err := scheduler.SchedulerCache.IsAssumedPod(pod)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to check whether pod %s/%s is assumed: %v", pod.Namespace, pod.Name, err))
		return false
	}
	return isAssumed
}

func (scheduler *Scheduler) frameworkForPod(pod *corev1.Pod) (*frameworkruntime.Framework, error) {
	fwk, ok := scheduler.Frameworks[pod.Spec.SchedulerName]
	if !ok {
		return nil, fmt.Errorf("profile not found for scheduler name %q", pod.Spec.SchedulerName)
	}
	return fwk, nil
}

func (scheduler *Scheduler) handleSchedulingFailure(fwk *frameworkruntime.Framework, podInfo *framework.QueuedPodInfo,
	err error, reason string, nominatingInfo *framework.NominatingInfo) {
	scheduler.Error(podInfo, err)
	if scheduler.PriorityQueue != nil {
		scheduler.PriorityQueue.AddNominatedPod(podInfo.PodInfo, nominatingInfo)
	}

	pod := podInfo.Pod
	fwk.EventRecorder().Eventf(pod, nil, corev1.EventTypeWarning, "FailedScheduling", "Scheduling", err.Error())
	if err := updatePod(scheduler.client, pod, &corev1.PodCondition{
		Type:    corev1.PodScheduled,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: err.Error(),
	}, nominatingInfo); err != nil {
		klog.ErrorS(err, "Error updating pod", "pod", klog.KObj(pod))
	}
}
func updatePod(client clientset.Interface, pod *corev1.Pod, condition *corev1.PodCondition, nominatingInfo *framework.NominatingInfo) error {
	klog.V(3).Infof("Updating pod condition for %s/%s to (%s==%s, Reason=%s)",
		pod.Namespace, pod.Name, condition.Type, condition.Status, condition.Reason)
	podStatusCopy := pod.Status.DeepCopy()
	// NominatedNodeName is updated only if we are trying to set it, and the value is
	// different from the existing one.
	nnnNeedsUpdate := nominatingInfo.Mode() == framework.ModeOverride && pod.Status.NominatedNodeName != nominatingInfo.NominatedNodeName
	if !podutil.UpdatePodCondition(podStatusCopy, condition) && !nnnNeedsUpdate {
		return nil
	}
	if nnnNeedsUpdate {
		podStatusCopy.NominatedNodeName = nominatingInfo.NominatedNodeName
	}

	return util.PatchPodStatus(client, pod, podStatusCopy)
}

// INFO:
func (scheduler *Scheduler) assume(assumed *corev1.Pod, host string) error {
	assumed.Spec.NodeName = host
	if err := scheduler.SchedulerCache.AssumePod(assumed); err != nil {
		klog.ErrorS(err, "Scheduler cache AssumePod failed")
		return err
	}
	// if "assumed" is a nominated pod, we should remove it from internal cache
	if scheduler.PriorityQueue != nil {
		scheduler.PriorityQueue.DeleteNominatedPodIfExists(assumed)
	}

	return nil
}

// INFO: filter 预选 and score 优选
//  (1) PreFilter/Filter
//  (2) PreScore/Score
func (scheduler *Scheduler) schedulePod(ctx context.Context, fwk *frameworkruntime.Framework,
	state *framework.CycleState, pod *corev1.Pod) (result ScheduleResult, err error) {
	if err := scheduler.SchedulerCache.UpdateSnapshot(scheduler.nodeInfoSnapshot); err != nil {
		return result, err
	}
	if scheduler.nodeInfoSnapshot.NumNodes() == 0 {
		return result, ErrNoNodesAvailable
	}

	// INFO: (1) PreFilter/Filter 预选出一批 feasibleNodes
	feasibleNodes, diagnosis, err := scheduler.findNodesThatFitPod(ctx, fwk, state, pod)
	if err != nil {
		return result, err
	}
	if len(feasibleNodes) == 0 {
		return result, &framework.FitError{
			Pod:         pod,
			NumAllNodes: scheduler.nodeInfoSnapshot.NumNodes(),
			Diagnosis:   diagnosis,
		}
	}
	// When only one node after predicate, just use it.
	if len(feasibleNodes) == 1 {
		return ScheduleResult{
			SuggestedHost:  feasibleNodes[0].Name,
			EvaluatedNodes: 1 + len(diagnosis.NodeToStatusMap),
			FeasibleNodes:  1,
		}, nil
	}

	// INFO: (2) PreScore/Score 优选出一批 nodes
	nodeScoreList, err := prioritizeNodes(ctx, fwk, state, pod, feasibleNodes)
	if err != nil {
		return result, err
	}
	host, err := selectHost(nodeScoreList)

	return ScheduleResult{
		SuggestedHost:  host,
		EvaluatedNodes: len(feasibleNodes) + len(diagnosis.NodeToStatusMap),
		FeasibleNodes:  len(feasibleNodes),
	}, err
}

// 找出分数最大的 node
func selectHost(nodeScoreList framework.NodeScoreList) (string, error) {
	if len(nodeScoreList) == 0 {
		return "", fmt.Errorf("empty priorityList")
	}
	maxScore := nodeScoreList[0].Score
	selected := nodeScoreList[0].Name
	cntOfMaxScore := 1
	for _, ns := range nodeScoreList[1:] {
		if ns.Score > maxScore {
			maxScore = ns.Score
			selected = ns.Name
			cntOfMaxScore = 1
		} else if ns.Score == maxScore {
			cntOfMaxScore++
			if rand.Intn(cntOfMaxScore) == 0 {
				// Replace the candidate with probability of 1/cntOfMaxScore
				selected = ns.Name
			}
		}
	}
	return selected, nil
}

// (1) 主要运行 PreFilter/Filter plugin, 过滤预选
func (scheduler *Scheduler) findNodesThatFitPod(ctx context.Context, fwk *frameworkruntime.Framework,
	state *framework.CycleState, pod *corev1.Pod) ([]*corev1.Node, framework.Diagnosis, error) {
	// Run "prefilter" plugins.
	diagnosis := framework.Diagnosis{
		NodeToStatusMap:      make(framework.NodeToStatusMap),
		UnschedulablePlugins: sets.NewString(),
	}
	preFilterResult, status := fwk.RunPreFilterPlugins(ctx, state, pod)
	allNodes, err := scheduler.nodeInfoSnapshot.NodeInfos().List()
	if err != nil {
		return nil, diagnosis, err
	}
	if !status.IsSuccess() {
		if !status.IsUnschedulable() {
			return nil, diagnosis, status.AsError()
		}
		for _, n := range allNodes {
			diagnosis.NodeToStatusMap[n.Node().Name] = status
		}
		// Status satisfying IsUnschedulable() gets injected into diagnosis.UnschedulablePlugins.
		if status.FailedPlugin() != "" {
			diagnosis.UnschedulablePlugins.Insert(status.FailedPlugin())
		}
		return nil, diagnosis, nil
	}

	// "NominatedNodeName" can potentially be set in a previous scheduling cycle as a result of preemption.
	// This node is likely the only candidate that will fit the pod, and hence we try it first before iterating over all nodes.
	// INFO: 这里优先选择 NominatedNodeName
	if len(pod.Status.NominatedNodeName) > 0 {
		feasibleNodes, err := scheduler.evaluateNominatedNode(ctx, pod, fwk, state, diagnosis)
		if err != nil {
			klog.ErrorS(err, "Evaluation failed on nominated node", "pod", klog.KObj(pod), "node", pod.Status.NominatedNodeName)
		}
		// Nominated node passes all the filters, scheduler is good to assign this node to the pod.
		if len(feasibleNodes) != 0 {
			return feasibleNodes, diagnosis, nil
		}
	}

	// nodes 过滤一遍 Filter plugins
	nodes := allNodes
	if !preFilterResult.AllNodes() {
		nodes = make([]*framework.NodeInfo, 0, len(preFilterResult.NodeNames))
		for n := range preFilterResult.NodeNames {
			nInfo, err := scheduler.nodeInfoSnapshot.NodeInfos().Get(n)
			if err != nil {
				return nil, diagnosis, err
			}
			nodes = append(nodes, nInfo)
		}
	}
	feasibleNodes, err := scheduler.findNodesThatPassFilters(ctx, fwk, state, pod, diagnosis, nodes)
	if err != nil {
		return nil, diagnosis, err
	}

	return feasibleNodes, diagnosis, nil
}

// NominatedNodeName 走一遍 Filter plugins
func (scheduler *Scheduler) evaluateNominatedNode(ctx context.Context, pod *corev1.Pod, fwk *frameworkruntime.Framework,
	state *framework.CycleState, diagnosis framework.Diagnosis) ([]*corev1.Node, error) {
	nnn := pod.Status.NominatedNodeName
	nodeInfo, err := scheduler.nodeInfoSnapshot.Get(nnn)
	if err != nil {
		return nil, err
	}
	node := []*framework.NodeInfo{nodeInfo}
	feasibleNodes, err := scheduler.findNodesThatPassFilters(ctx, fwk, state, pod, diagnosis, node)
	if err != nil {
		return nil, err
	}

	return feasibleNodes, nil
}

func (scheduler *Scheduler) findNodesThatPassFilters(
	ctx context.Context,
	fwk *frameworkruntime.Framework,
	state *framework.CycleState,
	pod *corev1.Pod,
	diagnosis framework.Diagnosis,
	nodes []*framework.NodeInfo) ([]*corev1.Node, error) {
	// Create feasible list with enough space to avoid growing it and allow assigning. 重新复制一遍
	numNodesToFind := scheduler.numFeasibleNodesToFind(int32(len(nodes)))
	feasibleNodes := make([]*corev1.Node, numNodesToFind)
	if !fwk.HasFilterPlugins() {
		length := len(nodes)
		for i := range feasibleNodes {
			feasibleNodes[i] = nodes[(scheduler.nextStartNodeIndex+i)%length].Node()
		}
		scheduler.nextStartNodeIndex = (scheduler.nextStartNodeIndex + len(feasibleNodes)) % length
		return feasibleNodes, nil
	}

	errCh := parallelize.NewErrorChannel()
	var statusesLock sync.Mutex
	var feasibleNodesLen int32
	ctx, cancel := context.WithCancel(ctx)
	checkNode := func(i int) {
		// We check the nodes starting from where we left off in the previous scheduling cycle,
		// this is to make sure all nodes have the same chance of being examined across pods.
		nodeInfo := nodes[(scheduler.nextStartNodeIndex+i)%len(nodes)]
		status := fwk.RunFilterPluginsWithNominatedPods(ctx, state, pod, nodeInfo)
		if status.Code() == framework.Error {
			errCh.SendErrorWithCancel(status.AsError(), cancel)
			return
		}
		if status.IsSuccess() {
			length := atomic.AddInt32(&feasibleNodesLen, 1)
			if length > numNodesToFind {
				cancel()
				atomic.AddInt32(&feasibleNodesLen, -1)
			} else {
				feasibleNodes[length-1] = nodeInfo.Node()
			}
		} else {
			statusesLock.Lock()
			diagnosis.NodeToStatusMap[nodeInfo.Node().Name] = status
			diagnosis.UnschedulablePlugins.Insert(status.FailedPlugin())
			statusesLock.Unlock()
		}
	}
	// 在寻找适合pod的node列表时，将开启16个（默认16个）goroutine 并行筛选，每个goroutine会各自负责所有节点中的一部分。
	fwk.Parallelizer().Until(ctx, len(nodes), checkNode)
	processedNodes := int(feasibleNodesLen) + len(diagnosis.NodeToStatusMap)
	scheduler.nextStartNodeIndex = (scheduler.nextStartNodeIndex + processedNodes) % len(nodes)

	feasibleNodes = feasibleNodes[:feasibleNodesLen]
	if err := errCh.ReceiveError(); err != nil {
		return nil, err
	}
	return feasibleNodes, nil
}

const (
	minFeasibleNodesToFind           = 100
	minFeasibleNodesPercentageToFind = 5
)

func (scheduler *Scheduler) numFeasibleNodesToFind(numAllNodes int32) (numNodes int32) {
	if numAllNodes < minFeasibleNodesToFind || scheduler.percentageOfNodesToScore >= 100 {
		return numAllNodes
	}

	adaptivePercentage := scheduler.percentageOfNodesToScore
	if adaptivePercentage <= 0 {
		basePercentageOfNodesToScore := int32(50)
		adaptivePercentage = basePercentageOfNodesToScore - numAllNodes/125
		if adaptivePercentage < minFeasibleNodesPercentageToFind {
			adaptivePercentage = minFeasibleNodesPercentageToFind
		}
	}
	numNodes = numAllNodes * adaptivePercentage / 100
	if numNodes < minFeasibleNodesToFind {
		return minFeasibleNodesToFind
	}

	return numNodes
}

// score 给 node 打分，优选
func prioritizeNodes(
	ctx context.Context,
	fwk *frameworkruntime.Framework,
	state *framework.CycleState,
	pod *corev1.Pod,
	nodes []*corev1.Node,
) (framework.NodeScoreList, error) {
	if !fwk.HasScorePlugins() {
		result := make(framework.NodeScoreList, 0, len(nodes))
		for i := range nodes {
			result = append(result, framework.NodeScore{
				Name:  nodes[i].Name,
				Score: 1,
			})
		}
		return result, nil
	}

	// Run PreScore plugins.
	preScoreStatus := fwk.RunPreScorePlugins(ctx, state, pod, nodes)
	if !preScoreStatus.IsSuccess() {
		return nil, preScoreStatus.AsError()
	}

	// Run the Score plugins.
	scoresMap, scoreStatus := fwk.RunScorePlugins(ctx, state, pod, nodes)
	if !scoreStatus.IsSuccess() {
		return nil, scoreStatus.AsError()
	}
	// Summarize all scores.
	result := make(framework.NodeScoreList, 0, len(nodes))
	for i := range nodes {
		result = append(result, framework.NodeScore{Name: nodes[i].Name, Score: 0})
		for j := range scoresMap {
			result[i].Score += scoresMap[j][i].Score
		}
	}

	return result, nil
}
