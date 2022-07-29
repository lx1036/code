// *********************************************************************
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-scheduling/scheduler_queues.md
// *********************************************************************

package queue

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"sync"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/framework"
	"k8s-lx1036/k8s/scheduler/pkg/internal/heap"
	"k8s-lx1036/k8s/scheduler/pkg/metrics"
	"k8s-lx1036/k8s/scheduler/pkg/util"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	DefaultPodInitialBackoffDuration time.Duration = 1 * time.Second

	DefaultPodMaxBackoffDuration time.Duration = 10 * time.Second

	DefaultPodMaxInUnschedulablePodsDuration time.Duration = 5 * time.Minute
)

// Events that trigger scheduler queue to change.
const (
	// Unknown event
	Unknown = "Unknown"
	// PodAdd is the event when a new pod is added to API server.
	PodAdd = "PodAdd"
	// NodeAdd is the event when a new node is added to the cluster.
	NodeAdd = "NodeAdd"
	// ScheduleAttemptFailure is the event when a schedule attempt fails.
	ScheduleAttemptFailure = "ScheduleAttemptFailure"
	// BackoffComplete is the event when a pod finishes backoff.
	BackoffComplete = "BackoffComplete"
	// AssignedPodUpdate is the event when a pod is updated that causes pods with matching affinity
	// terms to be more schedulable.
	AssignedPodUpdate = "AssignedPodUpdate"
	// AssignedPodDelete is the event when a pod is deleted that causes pods with matching affinity
	// terms to be more schedulable.
	AssignedPodDelete = "AssignedPodDelete"
	// PvAdd is the event when a persistent volume is added in the cluster.
	PvAdd = "PvAdd"
	// PvUpdate is the event when a persistent volume is updated in the cluster.
	PvUpdate = "PvUpdate"
	// PvcAdd is the event when a persistent volume claim is added in the cluster.
	PvcAdd = "PvcAdd"
	// PvcUpdate is the event when a persistent volume claim is updated in the cluster.
	PvcUpdate = "PvcUpdate"
	// StorageClassAdd is the event when a StorageClass is added in the cluster.
	StorageClassAdd = "StorageClassAdd"
	// ServiceAdd is the event when a service is added in the cluster.
	ServiceAdd = "ServiceAdd"
	// ServiceUpdate is the event when a service is updated in the cluster.
	ServiceUpdate = "ServiceUpdate"
	// ServiceDelete is the event when a service is deleted in the cluster.
	ServiceDelete = "ServiceDelete"
	// CSINodeAdd is the event when a CSI node is added in the cluster.
	CSINodeAdd = "CSINodeAdd"
	// CSINodeUpdate is the event when a CSI node is updated in the cluster.
	CSINodeUpdate = "CSINodeUpdate"
	// NodeSpecUnschedulableChange is the event when unschedulable node spec is changed.
	NodeSpecUnschedulableChange = "NodeSpecUnschedulableChange"
	// NodeAllocatableChange is the event when node allocatable is changed.
	NodeAllocatableChange = "NodeAllocatableChange"
	// NodeLabelsChange is the event when node label is changed.
	NodeLabelChange = "NodeLabelChange"
	// NodeTaintsChange is the event when node taint is changed.
	NodeTaintChange = "NodeTaintChange"
	// NodeConditionChange is the event when node condition is changed.
	NodeConditionChange = "NodeConditionChange"
)

func podInfoKeyFunc(obj interface{}) (string, error) {
	return cache.MetaNamespaceKeyFunc(obj.(*framework.QueuedPodInfo).Pod)
}

func WithClock(clock util.Clock) Option {
	return func(o *priorityQueueOptions) {
		o.clock = clock
	}
}

// WithPodInitialBackoffDuration sets pod initial backoff duration for PriorityQueue.
func WithPodInitialBackoffDuration(duration time.Duration) Option {
	return func(o *priorityQueueOptions) {
		o.podInitialBackoffDuration = duration
	}
}

// WithPodMaxBackoffDuration sets pod max backoff duration for PriorityQueue.
func WithPodMaxBackoffDuration(duration time.Duration) Option {
	return func(o *priorityQueueOptions) {
		o.podMaxBackoffDuration = duration
	}
}

// WithPodNominator sets pod nominator for PriorityQueue.
func WithPodNominator(pn *PodNominator) Option {
	return func(o *priorityQueueOptions) {
		o.podNominator = pn
	}
}

func WithClusterEventMap(m map[framework.ClusterEvent]sets.String) Option {
	return func(o *priorityQueueOptions) {
		o.clusterEventMap = m
	}
}

func MakeNextPodFunc(queue *PriorityQueue) func() *framework.QueuedPodInfo {
	return func() *framework.QueuedPodInfo {
		podInfo, err := queue.Pop()
		if err == nil {
			klog.V(4).InfoS("About to try and schedule pod", "pod", klog.KObj(podInfo.Pod))
			for plugin := range podInfo.UnschedulablePlugins {
				metrics.UnschedulableReason(plugin, podInfo.Pod.Spec.SchedulerName).Dec()
			}
			return podInfo
		}
		klog.ErrorS(err, "Error while retrieving next pod from scheduling queue")
		return nil
	}
}

// PriorityQueue 主要包含 activeQ/backoffQ(Heap实现的)，和 pod不可被调度的数据结构 unschedulableQ
type PriorityQueue struct {
	lock sync.RWMutex
	cond sync.Cond

	nsLister listersv1.NamespaceLister

	// activeQ is heap structure that scheduler actively looks at to find pods to
	// schedule. Head of heap is the highest priority pod.
	activeQ *heap.Heap
	// backoff time 最小置于队列最前
	podBackoffQ               *heap.Heap
	podInitialBackoffDuration time.Duration
	podMaxBackoffDuration     time.Duration
	// unschedulableQ holds pods that have been tried and determined unschedulable.
	unschedulablePods                 *UnschedulablePods
	podMaxInUnschedulablePodsDuration time.Duration
	// schedulingCycle represents sequence number of scheduling cycle and is incremented
	// when a pod is popped.
	schedulingCycle int64
	// moveRequestCycle caches the sequence number of scheduling cycle when we
	// received a move request. Unscheduable pods in and before this scheduling
	// cycle will be put back to activeQueue if we were trying to schedule them
	// when we received move request.
	moveRequestCycle int64

	PodNominator *PodNominator

	clusterEventMap map[framework.ClusterEvent]sets.String

	clock util.Clock

	stop chan struct{}

	// closed indicates that the queue is closed.
	// It is mainly used to let Pop() exit its control loop while waiting for an item.
	closed bool
}

type priorityQueueOptions struct {
	clock                             util.Clock
	podInitialBackoffDuration         time.Duration
	podMaxBackoffDuration             time.Duration
	podMaxInUnschedulablePodsDuration time.Duration
	podNominator                      *PodNominator
	clusterEventMap                   map[framework.ClusterEvent]sets.String
}

type Option func(*priorityQueueOptions)

var defaultPriorityQueueOptions = priorityQueueOptions{
	clock:                             util.RealClock{},
	podInitialBackoffDuration:         DefaultPodInitialBackoffDuration,
	podMaxBackoffDuration:             DefaultPodMaxBackoffDuration,
	podMaxInUnschedulablePodsDuration: DefaultPodMaxInUnschedulablePodsDuration,
}

// NewPriorityQueue creates a PriorityQueue object.
func NewPriorityQueue(
	lessFn framework.LessFunc,
	informerFactory informers.SharedInformerFactory,
	opts ...Option) *PriorityQueue {
	options := defaultPriorityQueueOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.podNominator == nil {
		options.podNominator = NewPodNominator(informerFactory.Core().V1().Pods().Lister())
	}

	pq := &PriorityQueue{
		PodNominator:                      options.podNominator,
		clock:                             options.clock,
		podInitialBackoffDuration:         options.podInitialBackoffDuration,
		podMaxBackoffDuration:             options.podMaxBackoffDuration,
		podMaxInUnschedulablePodsDuration: options.podMaxInUnschedulablePodsDuration,

		activeQ: heap.New(podInfoKeyFunc, func(podInfo1, podInfo2 interface{}) bool {
			pInfo1 := podInfo1.(*framework.QueuedPodInfo)
			pInfo2 := podInfo2.(*framework.QueuedPodInfo)
			return lessFn(pInfo1, pInfo2)
		}),
		unschedulablePods: newUnschedulablePodsMap(metrics.NewUnschedulablePodsRecorder()),
		moveRequestCycle:  -1,
		clusterEventMap:   options.clusterEventMap,

		stop: make(chan struct{}),
	}

	pq.cond.L = &pq.lock
	pq.podBackoffQ = heap.New(podInfoKeyFunc, pq.podsCompareBackoffCompleted)
	pq.nsLister = informerFactory.Core().V1().Namespaces().Lister()

	return pq
}

// Run 周期性的 flush backoffQ into activeQ(1s), unschedulePods into backoffQ or activeQ(30s)
func (p *PriorityQueue) Run() {
	go wait.Until(p.flushBackoffToActiveQueue, 1.0*time.Second, p.stop)
	go wait.Until(p.flushUnschedulablePodsToActiveOrBackoffQueue, 30*time.Second, p.stop)
}

func (p *PriorityQueue) podsCompareBackoffCompleted(podInfo1, podInfo2 interface{}) bool {
	pInfo1 := podInfo1.(*framework.QueuedPodInfo)
	pInfo2 := podInfo2.(*framework.QueuedPodInfo)
	return p.getBackoffTime(pInfo1).Before(p.getBackoffTime(pInfo2))
}

// getBackoffTime returns the time that podInfo completes backoff
func (p *PriorityQueue) getBackoffTime(podInfo *framework.QueuedPodInfo) time.Time {
	duration := p.calculateBackoffDuration(podInfo)
	backoffTime := podInfo.Timestamp.Add(duration)
	return backoffTime
}

// 周期性的 flush backoffQ into activeQ，逻辑还是比较简单的
func (p *PriorityQueue) flushBackoffToActiveQueue() {
	p.lock.Lock()
	defer p.lock.Unlock()

	broadcast := false
	for {
		rawPodInfo := p.podBackoffQ.Peek()
		if rawPodInfo == nil {
			break
		}
		pod := rawPodInfo.(*framework.QueuedPodInfo).Pod
		boTime := p.getBackoffTime(rawPodInfo.(*framework.QueuedPodInfo))
		if boTime.After(p.clock.Now()) { // 如果 backoff time 还没过期则跳过
			break
		}
		_, err := p.podBackoffQ.Pop() // delete pod from backoffQ
		if err != nil {
			klog.ErrorS(err, "Unable to pop pod from backoff queue despite backoff completion",
				"pod", klog.KObj(pod))
			break
		}
		p.activeQ.Add(rawPodInfo) // add into activeQ
		metrics.SchedulerQueueIncomingPods.WithLabelValues("active", BackoffComplete).Inc()
		broadcast = true
	}

	if broadcast {
		p.cond.Broadcast()
	}
}

// 周期性的 flush unschedulePods into activeQ or backoffQ，逻辑还是比较简单的
func (p *PriorityQueue) flushUnschedulablePodsToActiveOrBackoffQueue() {
	p.lock.Lock()
	defer p.lock.Unlock()

	// 该周期内，只有距离上次调度时间超过5分钟的 unschedulePods 才可以进行下一次调度
	var podsToMove []*framework.QueuedPodInfo
	currentTime := p.clock.Now()
	for _, podInfo := range p.unschedulablePods.podInfoMap {
		lastScheduleTime := podInfo.Timestamp
		if currentTime.Sub(lastScheduleTime) > p.podMaxInUnschedulablePodsDuration { // 大于 5min
			podsToMove = append(podsToMove, podInfo)
		}
	}

	if len(podsToMove) > 0 {
		p.movePodsToActiveOrBackoffQueue(podsToMove, UnschedulableTimeout)
	}
}

func (p *PriorityQueue) movePodsToActiveOrBackoffQueue(podInfoList []*framework.QueuedPodInfo, event framework.ClusterEvent) {
	for _, pInfo := range podInfoList {
		if len(pInfo.UnschedulablePlugins) != 0 && !p.podMatchesEvent(pInfo, event) {
			continue
		}

		pod := pInfo.Pod
		if p.isPodBackoff(pInfo) { // 如果 backoff time 还没过期，说明 pod 不是那么很旧，可以先放 backoffQ
			if err := p.podBackoffQ.Add(pInfo); err != nil { // unschedulableQ -> podBackoffQ
				klog.Errorf("Error adding pod %v to the backoff queue: %v", pod.Name, err)
			} else {
				metrics.SchedulerQueueIncomingPods.WithLabelValues("backoff", event.Label).Inc()
				p.unschedulablePods.delete(pod)
			}
		} else { // unschedulableQ -> activeQ，否则太旧的 pod 立刻放入 activeQ，尽快调度这个 pod
			if err := p.activeQ.Add(pInfo); err != nil {
				klog.Errorf("Error adding pod %v to the scheduling queue: %v", pod.Name, err)
			} else {
				metrics.SchedulerQueueIncomingPods.WithLabelValues("backoff", event.Label).Inc()
				p.unschedulablePods.delete(pod)
			}
		}
	}

	p.moveRequestCycle = p.schedulingCycle
	p.cond.Broadcast()
}

// 判断是不是 podBackoff pod
func (p *PriorityQueue) isPodBackoff(podInfo *framework.QueuedPodInfo) bool {
	return p.getBackoffTime(podInfo).After(p.clock.Now())
}

func (p *PriorityQueue) AddNominatedPod(pod *v1.Pod, nodeName string) {
	panic("implement me")
}

func (p *PriorityQueue) DeleteNominatedPodIfExists(pod *v1.Pod) {
	panic("implement me")
}

func (p *PriorityQueue) UpdateNominatedPod(oldPod, newPod *v1.Pod) {
	panic("implement me")
}

func (p *PriorityQueue) NominatedPodsForNode(nodeName string) []*v1.Pod {
	panic("implement me")
}

func (p *PriorityQueue) Update(oldPod, newPod *v1.Pod) error {
	panic("implement me")
}

func (p *PriorityQueue) Delete(pod *v1.Pod) error {
	panic("implement me")
}

func (p *PriorityQueue) AssignedPodUpdated(pod *v1.Pod) {
	panic("implement me")
}

func (p *PriorityQueue) PendingPods() []*v1.Pod {
	panic("implement me")
}

func (p *PriorityQueue) NumUnschedulablePods() int {
	panic("implement me")
}

// newQueuedPodInfo builds a QueuedPodInfo object.
func (p *PriorityQueue) newQueuedPodInfo(pod *v1.Pod) *framework.QueuedPodInfo {
	now := p.clock.Now()
	return &framework.QueuedPodInfo{
		Pod:                     pod,
		Timestamp:               now,
		InitialAttemptTimestamp: now,
	}
}
func newQueuedPodInfoNoTimestamp(pod *v1.Pod) *framework.QueuedPodInfo {
	return &framework.QueuedPodInfo{
		Pod: pod,
	}
}

// add pod to activeQ
func (p *PriorityQueue) Add(pod *v1.Pod) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	podInfo := p.newQueuedPodInfo(pod)
	if err := p.activeQ.Add(podInfo); err != nil {
		klog.Errorf("Error adding pod %s/%s to the scheduling queue: %v", pod.Namespace, pod.Name, err)
		return err
	}
	if p.unschedulablePods.get(pod) != nil {
		klog.Errorf("Error: pod %s/%s is already in the unschedulable queue.", pod.Namespace, pod.Name)
		p.unschedulablePods.delete(pod)
	}
	// Delete pod from backoffQ if it is backing off
	if err := p.podBackoffQ.Delete(podInfo); err == nil {
		klog.Errorf("Error: pod %s/%s is already in the podBackoff queue.", pod.Namespace, pod.Name)
	}

	p.PodNominator.AddNominatedPod(pod, "")
	p.cond.Broadcast()

	return nil
}

func (p *PriorityQueue) AddUnschedulableIfNotPresent(pInfo *framework.QueuedPodInfo, podSchedulingCycle int64) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	pod := pInfo.Pod
	if p.unschedulablePods.get(pod) != nil {
		return fmt.Errorf("pod: %s/%s is already present in unschedulable queue", pod.Namespace, pod.Name)
	}

	// Refresh the timestamp since the pod is re-added.
	pInfo.Timestamp = p.clock.Now()
	if _, exists, _ := p.activeQ.Get(pInfo); exists {
		return fmt.Errorf("pod: %s/%s is already present in the active queue", pod.Namespace, pod.Name)
	}
	if _, exists, _ := p.podBackoffQ.Get(pInfo); exists {
		return fmt.Errorf("pod %s/%s is already present in the backoff queue", pod.Namespace, pod.Name)
	}

	// If a move request has been received, move it to the BackoffQ, otherwise move it to unschedulableQ.
	if p.moveRequestCycle >= podSchedulingCycle {
		if err := p.podBackoffQ.Add(pInfo); err != nil {
			return fmt.Errorf("error adding pod %v to the backoff queue: %v", pod.Name, err)
		}
	} else {
		p.unschedulablePods.addOrUpdate(pInfo)
	}

	p.PodNominator.AddNominatedPod(pod, "")
	return nil
}

const queueClosed = "scheduling queue is closed"

// 最大堆activeQ中pop一个pod出来，没有则一直block等待，同时p.schedulingCycle++
// Pop() 函数会阻塞，这点很重要！！！
func (p *PriorityQueue) Pop() (*framework.QueuedPodInfo, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for p.activeQ.Len() == 0 {
		// When the queue is empty, invocation of Pop() is blocked until new item is enqueued.
		// When Close() is called, the p.closed is set and the condition is broadcast,
		// which causes this loop to continue and return from the Pop().
		if p.closed {
			return nil, fmt.Errorf(queueClosed)
		}
		p.cond.Wait()
	}

	obj, err := p.activeQ.Pop()
	if err != nil {
		return nil, err
	}
	pInfo := obj.(*framework.QueuedPodInfo)
	pInfo.Attempts++
	p.schedulingCycle++
	return pInfo, err
}

// 该pod会把unschedulableQ中与其affinity匹配的pod放到activeQ中
// 这样可以使得两个亲和性pod优先被调度起来
func (p *PriorityQueue) AssignedPodAdded(pod *v1.Pod) {
	p.lock.Lock()
	p.movePodsToActiveOrBackoffQueue(p.getUnschedulablePodsWithMatchingAffinityTerm(pod), AssignedPodAdd)
	p.lock.Unlock()
}

// 从 unschedulableQ 中寻找pods，该pods需要match到输入的pod affinity
func (p *PriorityQueue) getUnschedulablePodsWithMatchingAffinityTerm(pod *v1.Pod) []*framework.QueuedPodInfo {
	var podsToMove []*framework.QueuedPodInfo
	for _, pInfo := range p.unschedulablePods.podInfoMap {
		up := pInfo.Pod
		terms := util.GetPodAffinityTerms(up.Spec.Affinity)
		for _, term := range terms {
			namespaces := util.GetNamespacesFromPodAffinityTerm(up, &term)
			selector, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
			if err != nil {
				klog.Errorf("Error getting label selectors for pod: %v.", up.Name)
			}
			if util.PodMatchesTermsNamespaceAndSelector(pod, namespaces, selector) {
				podsToMove = append(podsToMove, pInfo)
				break
			}
		}
	}

	return podsToMove
}

// 把 unschedulableQ 和 podBackoffQ 全部 move 到 activeQ
func (p *PriorityQueue) MoveAllToActiveOrBackoffQueue(event framework.ClusterEvent) {
	p.lock.Lock()
	defer p.lock.Unlock()
	unschedulablePods := make([]*framework.QueuedPodInfo, 0, len(p.unschedulablePods.podInfoMap))
	for _, pInfo := range p.unschedulablePods.podInfoMap {
		unschedulablePods = append(unschedulablePods, pInfo)
	}
	p.movePodsToActiveOrBackoffQueue(unschedulablePods, event)
}

// p.podInitialBackoffDuration 每次翻倍，次数不能超过podInfo.Attempts，也不能超过最大值
func (p *PriorityQueue) calculateBackoffDuration(podInfo *framework.QueuedPodInfo) time.Duration {
	duration := p.podInitialBackoffDuration
	for i := 1; i < podInfo.Attempts; i++ {
		duration = duration * 2
		if duration > p.podMaxBackoffDuration {
			return p.podMaxBackoffDuration // 最大 10s
		}
	}
	return duration
}

func (p *PriorityQueue) SchedulingCycle() int64 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.schedulingCycle
}

func (p *PriorityQueue) Close() {
	panic("implement me")
}
