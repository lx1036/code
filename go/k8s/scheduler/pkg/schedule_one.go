package pkg

import (
	"context"
	"fmt"
	"math/rand"

	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/metrics"
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
	// pod could be nil when schedulerQueue is closed
	if podInfo == nil || podInfo.Pod == nil {
		return
	}
	pod := podInfo.Pod
	fwk, err := scheduler.frameworkForPod(pod)
	if err != nil {
		// This shouldn't happen, because we only accept for scheduling the pods
		// which specify a scheduler name that matches one of the profiles.
		klog.Error(err)
		return
	}
	if scheduler.skipPodSchedule(fwk, pod) {
		return
	}

	klog.Infof("Attempting to schedule pod: %v/%v", pod.Namespace, pod.Name)

	// INFO: 由 schedule algo 来串起来并实际执行各个plugins
	//start := time.Now()
	state := framework.NewCycleState()
	// INFO: 这里逻辑只有10%概率记录 plugin metrics
	state.SetRecordPluginMetrics(rand.Intn(100) < pluginMetricsSamplePercent)
	schedulingCycleCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	scheduleResult, err := scheduler.schedulePod(schedulingCycleCtx, fwk, state, pod)
	if err != nil {
		nominatedNode := ""
		// INFO: 如果pod调度失败，则调用 PostFilter plugin 进行抢占
		if fitError, ok := err.(*core.FitError); ok {
			if !framework.HasPostFilterPlugins() {
				klog.V(3).Infof("No PostFilter plugins are registered, so no preemption will be performed.")
			} else {
				// INFO: PostFilter plugin 其实就是 defaultpreemption.Name plugin，运行 preemption plugin
				result, status := framework.RunPostFilterPlugins(ctx, state, pod, fitError.FilteredNodesStatuses)
				if status.Code() == framework.Error {
					klog.Errorf("Status after running PostFilter plugins for pod %v/%v: %v", pod.Namespace, pod.Name, status)
				} else {
					klog.V(5).Infof("Status after running PostFilter plugins for pod %v/%v: %v", pod.Namespace, pod.Name, status)
				}
				if status.IsSuccess() && result != nil {
					// INFO: 如果抢占成功，则去更新 pod.Status.NominatedNodeName，但是这次调度周期不会立刻更新 pod.Spec.nodeName，
					// 等待下次调度周期去调度。同时，下次调度周期时 pod.Spec.nodeName 未必就是 pod.Status.NominatedNodeName 这个 node
					// 可以去看 k8s.io/api/core/v1/types.go::NominatedNodeName 字段定义描述
					nominatedNode = result.NominatedNodeName
				}
			}
			// metrics
		} else if err == core.ErrNoNodesAvailable {

		} else {
			klog.ErrorS(err, "Error selecting node for pod", "pod", klog.KObj(pod))
		}

		// INFO: 更新 pod.Status.NominatedNodeName，以及更新 pod.Status.Conditions 便于展示信息
		scheduler.recordSchedulingFailure(framework, podInfo, err, corev1.PodReasonUnschedulable, nominatedNode)

		return
	}

	// Run the Reserve method of reserve plugins.
	if sts := fwk.RunReservePluginsReserve(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost); !sts.IsSuccess() {
		metrics.PodScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
		// trigger un-reserve to clean up state associated with the reserved Pod
		fwk.RunReservePluginsUnreserve(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if forgetErr := sched.Cache.ForgetPod(assumedPod); forgetErr != nil {
			klog.ErrorS(forgetErr, "Scheduler cache ForgetPod failed")
		}
		sched.handleSchedulingFailure(fwk, assumedPodInfo, sts.AsError(), SchedulerError, clearNominatedNode)
		return
	}

	// Run "permit" plugins.
	runPermitStatus := fwk.RunPermitPlugins(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
	if runPermitStatus.Code() != framework.Wait && !runPermitStatus.IsSuccess() {

	}

	// 启动goroutine执行bind操作
	go func() {
		bindingCycleCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		waitOnPermitStatus := fwk.WaitOnPermit(bindingCycleCtx, assumedPod)
		if !waitOnPermitStatus.IsSuccess() {

		}
		// Run "prebind" plugins.
		preBindStatus := fwk.RunPreBindPlugins(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if !preBindStatus.IsSuccess() {

		}

		err := scheduler.bind(bindingCycleCtx, prof, assumedPod, scheduleResult.SuggestedHost, state)
		if err != nil {

		} else {

			// Run "postbind" plugins.
			fwk.RunPostBindPlugins(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		}
	}()
}

// skipPodSchedule returns true if we could skip scheduling the pod for specified cases.
func (scheduler *Scheduler) skipPodSchedule(prof *frameworkruntime.Framework, pod *corev1.Pod) bool {
	// ...
	return false
	// 存入PriorityQueue
	scheduler.PriorityQueue.AssignedPodAdded(pod)
	return false
}

func (scheduler *Scheduler) frameworkForPod(pod *corev1.Pod) (*frameworkruntime.Framework, error) {
	framework, ok := scheduler.Frameworks[pod.Spec.SchedulerName]
	if !ok {
		return nil, fmt.Errorf("profile not found for scheduler name %q", pod.Spec.SchedulerName)
	}
	return framework, nil
}

// INFO: filter 预选 and score 优选
func (scheduler *Scheduler) schedulePod(ctx context.Context, fwk *frameworkruntime.Framework,
	state *framework.CycleState, pod *corev1.Pod) (result ScheduleResult, err error) {

	feasibleNodes, diagnosis, err := scheduler.findNodesThatFitPod(ctx, fwk, state, pod)
	if err != nil {
		return result, err
	}
	if len(feasibleNodes) == 0 {
		return result, &framework.FitError{
			Pod:         pod,
			NumAllNodes: sched.nodeInfoSnapshot.NumNodes(),
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

func (scheduler *Scheduler) bind(ctx context.Context, fwk frameworkruntime.Framework,
	assumed *corev1.Pod, targetNode string, state *framework.CycleState) (err error) {

	bindStatus := fwk.RunBindPlugins(ctx, state, assumed, targetNode)

}
