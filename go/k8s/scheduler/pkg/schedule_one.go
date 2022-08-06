package pkg

import (
	"context"
	"fmt"
	v1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/api/core/v1"
	internalqueue "k8s-lx1036/k8s/scheduler/pkg/internal/queue"
	"k8s.io/kubernetes/pkg/scheduler/metrics"
	"math/rand"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
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

	// At the end of a successful scheduling cycle, pop and move up Pods if needed.
	if len(podsToActivate.Map) != 0 {
		scheduler.PriorityQueue.Activate(podsToActivate.Map)
		// Clear the entries after activation.
		podsToActivate.Map = make(map[string]*corev1.Pod)
	}

	// INFO: (6) PreBind(有 VolumeBinding)/Bind(只有 DefaultBind)/PostBind(暂无)
	go func() {
		bindingCycleCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		waitOnPermitStatus := fwk.WaitOnPermit(bindingCycleCtx, assumedPod)
		if !waitOnPermitStatus.IsSuccess() {
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

		err := scheduler.bind(bindingCycleCtx, fwk, assumedPod, scheduleResult.SuggestedHost, state)
		if err != nil {
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

	priorityList, err := prioritizeNodes(ctx, fwk, state, pod, feasibleNodes)
	if err != nil {
		return result, err
	}
	host, err := selectHost(priorityList)

}

// (1) 主要运行 PreFilter/Filter plugin, 过滤预选
func (scheduler *Scheduler) findNodesThatFitPod(ctx context.Context, fwk *frameworkruntime.Framework,
	state *framework.CycleState, pod *corev1.Pod) ([]*corev1.Node, framework.NodeToStatusMap, error) {
	filteredNodesStatuses := make(framework.NodeToStatusMap)

	// Run "prefilter" plugins.
	s := fwk.RunPreFilterPlugins(ctx, state, pod)

	// Run "filter" plugins.
	feasibleNodes, err := scheduler.findNodesThatPassFilters(ctx, fwk, state, pod, filteredNodesStatuses)
	if err != nil {
		return nil, nil, err
	}

	return feasibleNodes, diagnosis, nil
}

func (scheduler *Scheduler) findNodesThatPassFilters(
	ctx context.Context,
	fwk *frameworkruntime.Framework,
	state *framework.CycleState,
	pod *corev1.Pod,
	diagnosis framework.Diagnosis,
	nodes []*framework.NodeInfo) ([]*corev1.Node, error) {
	if !fwk.HasFilterPlugins() {

	}

	checkNode := func(i int) {
		// We check the nodes starting from where we left off in the previous scheduling cycle,
		// this is to make sure all nodes have the same chance of being examined across pods.
		nodeInfo := nodes[(scheduler.nextStartNodeIndex+i)%len(nodes)]
		status := fwk.RunFilterPluginsWithNominatedPods(ctx, state, pod, nodeInfo)
	}
	// Stops searching for more nodes once the configured number of feasible nodes
	// are found.
	fwk.Parallelizer().Until(ctx, len(nodes), checkNode)

}

// score 给 node 打分，优选
func prioritizeNodes(
	ctx context.Context,
	fwk *frameworkruntime.Framework,
	state *framework.CycleState,
	pod *corev1.Pod,
	nodes []*corev1.Node,
) (framework.NodeScoreList, error) {

}

func (scheduler *Scheduler) bind(ctx context.Context, fwk *frameworkruntime.Framework,
	assumed *corev1.Pod, targetNode string, state *framework.CycleState) (err error) {

	bindStatus := fwk.RunBindPlugins(ctx, state, assumed, targetNode)

}
