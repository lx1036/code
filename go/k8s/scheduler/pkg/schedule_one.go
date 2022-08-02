package pkg

import (
	"context"
	"fmt"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	corev1 "k8s.io/api/core/v1"
	"math/rand"

	"k8s-lx1036/k8s/scheduler/pkg/framework"

	"k8s-lx1036/k8s/scheduler/pkg/core"
	"k8s.io/klog/v2"
)

func (scheduler *Scheduler) scheduleOne(ctx context.Context) {
	podInfo := scheduler.NextPod()
	// pod could be nil when schedulerQueue is closed
	if podInfo == nil || podInfo.Pod == nil {
		return
	}
	pod := podInfo.Pod
	prof, err := scheduler.profileForPod(pod)
	if err != nil {
		// This shouldn't happen, because we only accept for scheduling the pods
		// which specify a scheduler name that matches one of the profiles.
		klog.Error(err)
		return
	}
	if scheduler.skipPodSchedule(prof, pod) {
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
	scheduleResult, err := scheduler.Algorithm.Schedule(schedulingCycleCtx, prof, state, pod)
	scheduleResult, err := scheduler.SchedulePod(schedulingCycleCtx, fwk, state, pod)
	if err != nil {
		nominatedNode := ""
		// INFO: 如果pod调度失败，则调用 PostFilter plugin 进行抢占
		if fitError, ok := err.(*core.FitError); ok {
			if !prof.HasPostFilterPlugins() {
				klog.V(3).Infof("No PostFilter plugins are registered, so no preemption will be performed.")
			} else {
				// INFO: PostFilter plugin 其实就是 defaultpreemption.Name plugin，运行 preemption plugin
				result, status := prof.RunPostFilterPlugins(ctx, state, pod, fitError.FilteredNodesStatuses)
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
		scheduler.recordSchedulingFailure(prof, podInfo, err, corev1.PodReasonUnschedulable, nominatedNode)

		return
	}

	// Run "permit" plugins.
	runPermitStatus := prof.RunPermitPlugins(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
	if runPermitStatus.Code() != framework.Wait && !runPermitStatus.IsSuccess() {

	}

	// 启动goroutine执行bind操作
	go func() {
		waitOnPermitStatus := prof.WaitOnPermit(bindingCycleCtx, assumedPod)
		if !waitOnPermitStatus.IsSuccess() {

		}
		// Run "prebind" plugins.
		preBindStatus := prof.RunPreBindPlugins(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if !preBindStatus.IsSuccess() {

		}

		err := scheduler.bind(bindingCycleCtx, prof, assumedPod, scheduleResult.SuggestedHost, state)
		if err != nil {

		} else {

			// Run "postbind" plugins.
			prof.RunPostBindPlugins(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
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

func (scheduler *Scheduler) profileForPod(pod *corev1.Pod) (*frameworkruntime.Framework, error) {
	prof, ok := scheduler.Frameworks[pod.Spec.SchedulerName]
	if !ok {
		return nil, fmt.Errorf("profile not found for scheduler name %q", pod.Spec.SchedulerName)
	}
	return prof, nil
}

func (scheduler *Scheduler) schedulePod(ctx context.Context, fwk framework.Framework,
	state *framework.CycleState, pod *corev1.Pod) (result ScheduleResult, err error) {

}
