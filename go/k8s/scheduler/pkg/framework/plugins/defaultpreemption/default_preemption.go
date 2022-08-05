package defaultpreemption

import (
	"context"
	"errors"
	"fmt"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sync"

	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	"k8s-lx1036/k8s/scheduler/pkg/metrics"
	"k8s-lx1036/k8s/scheduler/pkg/util"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corelisters "k8s.io/client-go/listers/core/v1"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

// INFO: 抢占preemption plugin

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

func (pl *DefaultPreemption) PostFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod,
	m framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	defer func() {
		metrics.PreemptionAttempts.Inc()
	}()

	nominatedNodeName, err := pl.preempt(ctx, state, pod, m)
	if err != nil {
		return nil, framework.NewStatus(framework.Error, err.Error())
	}
	if nominatedNodeName == "" {
		return nil, framework.NewStatus(framework.Unschedulable)
	}

	return &framework.PostFilterResult{NominatedNodeName: nominatedNodeName}, framework.NewStatus(framework.Success)
}

func (pl *DefaultPreemption) preempt(ctx context.Context, state *framework.CycleState, pod *v1.Pod,
	m framework.NodeToStatusMap) (string, error) {
	nodeLister := pl.framework.SnapshotSharedLister().NodeInfos()

	podNamespace, podName := pod.Namespace, pod.Name
	pod, err := pl.podLister.Pods(pod.Namespace).Get(pod.Name)
	if err != nil {
		klog.ErrorS(err, "Getting the updated preemptor pod object", "pod", klog.KRef(podNamespace, podName))
		return nil, framework.AsStatus(err)
	}

	// INFO:(1) Ensure the preemptor is eligible to preempt other pods.
	if !pl.PodEligibleToPreemptOthers(pod, m[pod.Status.NominatedNodeName]) {
		klog.V(5).Infof("Pod %v/%v is not eligible for more preemption.", pod.Namespace, pod.Name)
		return "", nil
	}

	// INFO:(2) Find all preemption candidates.
	candidates, nodeToStatusMap, err := pl.findCandidates(ctx, pod, m, state)
	if err != nil && len(candidates) == 0 {
		return nil, framework.AsStatus(err)
	}

	// INFO: (3) Find the best candidate.
	bestCandidate := pl.SelectCandidate(candidates)
	if bestCandidate == nil || len(bestCandidate.Name()) == 0 {
		return nil, framework.NewStatus(framework.Unschedulable, "no candidate node for preemption")
	}

	// INFO: (4) Perform preparation work before nominating the selected candidate.
	if status := pl.prepareCandidate(bestCandidate, pod, pl.PluginName); !status.IsSuccess() {
		return nil, status
	}

	return framework.NewPostFilterResultWithNominatedNode(bestCandidate.Name()), framework.NewStatus(framework.Success)
}

// PodEligibleToPreemptOthers INFO: @see https://kubernetes.io/zh/docs/concepts/configuration/pod-priority-preemption/#non-preempting-priority-class
func (pl *DefaultPreemption) PodEligibleToPreemptOthers(pod *v1.Pod, nominatedNodeStatus *framework.Status) bool {
	// INFO: 非抢占式的 pod 不需要抢占，返回 false
	if pod.Spec.PreemptionPolicy != nil && *pod.Spec.PreemptionPolicy == v1.PreemptNever {
		klog.V(5).Infof("Pod %v/%v is not eligible for preemption because it has a preemptionPolicy of %v", pod.Namespace, pod.Name, v1.PreemptNever)
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

func (pl *DefaultPreemption) findCandidates(ctx context.Context, pod *v1.Pod,
	m framework.NodeToStatusMap, state *framework.CycleState) ([]*Candidate, framework.NodeToStatusMap, error) {
	allNodes, err := pl.framework.SnapshotSharedLister().NodeInfos().List()
	if err != nil {
		return nil, nil, err
	}
	if len(allNodes) == 0 {
		return nil, nil, errors.New("no nodes available")
	}
	potentialNodes, unschedulableNodeStatus := nodesWherePreemptionMightHelp(allNodes, m)
	if len(potentialNodes) == 0 {
		klog.V(3).InfoS("Preemption will not help schedule pod on any node", "pod", klog.KObj(pod))
		// In this case, we should clean-up any existing nominated node name of the pod.
		if err := util.ClearNominatedNodeName(pl.framework.ClientSet(), pod); err != nil {
			klog.ErrorS(err, "Cannot clear 'NominatedNodeName' field of pod", "pod", klog.KObj(pod))
			// We do not return as this error is not critical.
		}
		return nil, unschedulableNodeStatus, nil
	}

	offset, numCandidates := pl.GetOffsetAndNumCandidates(int32(len(potentialNodes)))
	candidates, nodeStatuses, err := pl.DryRunPreemption(ctx, pod, potentialNodes, offset, numCandidates, state)
	for node, nodeStatus := range unschedulableNodeStatus {
		nodeStatuses[node] = nodeStatus
	}
	return candidates, nodeStatuses, err
}

func (pl *DefaultPreemption) DryRunPreemption(ctx context.Context, pod *v1.Pod, potentialNodes []*framework.NodeInfo,
	offset int32, numCandidates int32, state *framework.CycleState) ([]Candidate, framework.NodeToStatusMap, error) {
	// INFO: 为了高效率，这里通过多个并发去处理 potentialNodes，而不是一个个去处理 node
	var statusesLock sync.Mutex
	var errs []error
	nodeStatuses := make(framework.NodeToStatusMap)
	nonViolatingCandidates := newCandidateList(numCandidates)
	violatingCandidates := newCandidateList(numCandidates)
	parallelCtx, cancel := context.WithCancel(ctx)
	checkNode := func(i int) {
		nodeInfoCopy := potentialNodes[(int(offset)+i)%len(potentialNodes)].Clone()
		stateCopy := state.Clone()
		pods, numPDBViolations, status := pl.SelectVictimsOnNode(ctx, stateCopy, pod, nodeInfoCopy)
		if status.IsSuccess() && len(pods) != 0 {
			victims := extenderv1.Victims{
				Pods:             pods,
				NumPDBViolations: int64(numPDBViolations),
			}
			c := &Candidate{
				victims: &victims,
				name:    nodeInfoCopy.Node().Name,
			}
			if numPDBViolations == 0 {
				nonViolatingCandidates.add(c)
			} else {
				violatingCandidates.add(c)
			}
			nvcSize, vcSize := nonViolatingCandidates.size(), violatingCandidates.size()
			if nvcSize > 0 && nvcSize+vcSize >= numCandidates {
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
	return append(nonViolatingCandidates.get(), violatingCandidates.get()...), nodeStatuses, utilerrors.NewAggregate(errs)
}

// SelectCandidate 选择 best-fit candidate
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

}

// prepareCandidate does some preparation work before nominating the selected candidate:
// - Evict the victim pods
// - Reject the victim pods if they are in waitingPod map
// - Clear the low-priority pods' nominatedNodeName status if needed
func (pl *DefaultPreemption) prepareCandidate(c Candidate, pod *v1.Pod, pluginName string) *framework.Status {
	for _, victim := range c.Victims().Pods {
		// If the victim is a WaitingPod, send a reject message to the PermitPlugin.
		// Otherwise we should delete the victim.
		if waitingPod := pl.GetWaitingPod(victim.UID); waitingPod != nil {
			waitingPod.Reject(pluginName, "preempted")
		} else if err := util.DeletePod(pl.framework.ClientSet(), victim); err != nil {
			klog.ErrorS(err, "Preempting pod", "pod", klog.KObj(victim), "preemptor", klog.KObj(pod))
			return framework.AsStatus(err)
		}
		pl.framework.EventRecorder().Eventf(victim, pod, v1.EventTypeNormal, "Preempted", "Preempting", "Preempted by %v/%v on node %v",
			pod.Namespace, pod.Name, c.Name())
	}

	nominatedPods := getLowerPriorityNominatedPods(pl.framework, pod, c.Name())
	if err := util.ClearNominatedNodeName(pl.framework.ClientSet(), nominatedPods...); err != nil {
		klog.ErrorS(err, "Cannot clear 'NominatedNodeName' field")
		// We do not return as this error is not critical.
	}

	return nil
}
