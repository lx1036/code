package queue

import (
	"reflect"
	"testing"
	"time"

	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/queuesort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/clock"
)

func newDefaultQueueSort() framework.LessFunc {
	sort := &queuesort.PrioritySort{}
	return sort.Less // pod Priority 高则排在最前面，最大堆
}

var lowPriority, midPriority, highPriority = int32(0), int32(100), int32(1000)
var mediumPriority = (lowPriority + highPriority) / 2

var highPriorityPod, highPriNominatedPod, medPriorityPod, unschedulablePod = v1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "hpp",
		Namespace: "ns1",
		UID:       "hppns1",
	},
	Spec: v1.PodSpec{
		Priority: &highPriority,
	},
},
	v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hpp",
			Namespace: "ns1",
			UID:       "hppns1",
		},
		Spec: v1.PodSpec{
			Priority: &highPriority,
		},
		Status: v1.PodStatus{
			NominatedNodeName: "node1",
		},
	},
	v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mpp",
			Namespace: "ns2",
			UID:       "mppns2",
			Annotations: map[string]string{
				"annot2": "val2",
			},
		},
		Spec: v1.PodSpec{
			Priority: &mediumPriority,
		},
		Status: v1.PodStatus{
			NominatedNodeName: "node1",
		},
	},
	v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "up",
			Namespace: "ns1",
			UID:       "upns1",
			Annotations: map[string]string{
				"annot2": "val2",
			},
		},
		Spec: v1.PodSpec{
			Priority: &lowPriority,
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodScheduled,
					Status: v1.ConditionFalse,
					Reason: v1.PodReasonUnschedulable,
				},
			},
			NominatedNodeName: "node1",
		},
	}

func TestPriorityQueueAdd(test *testing.T) {
	q := NewPriorityQueue(newDefaultQueueSort())
	if err := q.Add(&medPriorityPod); err != nil {
		test.Errorf("add failed: %v", err)
	}
	if err := q.Add(&unschedulablePod); err != nil {
		test.Errorf("add failed: %v", err)
	}
	if err := q.Add(&highPriorityPod); err != nil {
		test.Errorf("add failed: %v", err)
	}

	expectedNominatedPods := &nominatedPodMap{
		nominatedPodToNode: map[types.UID]string{
			medPriorityPod.UID:   "node1",
			unschedulablePod.UID: "node1",
		},
		nominatedPods: map[string][]*v1.Pod{
			"node1": {&medPriorityPod, &unschedulablePod},
		},
	}
	if !reflect.DeepEqual(q.PodNominator, expectedNominatedPods) {
		test.Errorf("Unexpected nominated map after adding pods. Expected: %v, got: %v", expectedNominatedPods, q.PodNominator)
	}

	if p, err := q.Pop(); err != nil || p.Pod != &highPriorityPod {
		test.Errorf("Expected: %v after Pop, but got: %v", highPriorityPod.Name, p.Pod.Name)
	}
	if p, err := q.Pop(); err != nil || p.Pod != &medPriorityPod {
		test.Errorf("Expected: %v after Pop, but got: %v", medPriorityPod.Name, p.Pod.Name)
	}
	if p, err := q.Pop(); err != nil || p.Pod != &unschedulablePod {
		test.Errorf("Expected: %v after Pop, but got: %v", unschedulablePod.Name, p.Pod.Name)
	}
	if len(q.PodNominator.nominatedPods["node1"]) != 2 {
		test.Errorf("Expected medPriorityPod and unschedulablePod to be still present in nomindatePods: %v", q.PodNominator.nominatedPods["node1"])
	}
}

func TestPriorityQueueAddWithReverse(test *testing.T) {
	q := NewPriorityQueue(newDefaultQueueSort())
	if err := q.Add(&medPriorityPod); err != nil {
		test.Errorf("add failed: %v", err)
	}
	if err := q.Add(&highPriorityPod); err != nil {
		test.Errorf("add failed: %v", err)
	}
	if p, err := q.Pop(); err != nil || p.Pod != &highPriorityPod { // 直接相等比较，没用reflect.DeepEqual
		test.Errorf("Expected: %v after Pop, but got: %v", highPriorityPod.Name, p.Pod.Name)
	}
	if p, err := q.Pop(); err != nil || p.Pod != &medPriorityPod {
		test.Errorf("Expected: %v after Pop, but got: %v", medPriorityPod.Name, p.Pod.Name)
	}
}

func getUnschedulablePod(p *PriorityQueue, pod *v1.Pod) *v1.Pod {
	pInfo := p.unschedulableQ.get(pod)
	if pInfo != nil {
		return pInfo.Pod
	}
	return nil
}

func TestPriorityQueueAddUnschedulableIfNotPresent(t *testing.T) {
	q := NewPriorityQueue(newDefaultQueueSort())
	_ = q.Add(&highPriNominatedPod)
	_ = q.AddUnschedulableIfNotPresent(newQueuedPodInfoNoTimestamp(&highPriNominatedPod), q.SchedulingCycle()) // Must not add anything.
	_ = q.AddUnschedulableIfNotPresent(newQueuedPodInfoNoTimestamp(&unschedulablePod), q.SchedulingCycle())
	expectedNominatedPods := &nominatedPodMap{
		nominatedPodToNode: map[types.UID]string{
			unschedulablePod.UID:    "node1",
			highPriNominatedPod.UID: "node1",
		},
		nominatedPods: map[string][]*v1.Pod{
			"node1": {&highPriNominatedPod, &unschedulablePod},
		},
	}
	if !reflect.DeepEqual(q.PodNominator, expectedNominatedPods) {
		t.Errorf("Unexpected nominated map after adding pods. Expected: %v, got: %v", expectedNominatedPods, q.PodNominator)
	}
	if p, err := q.Pop(); err != nil || p.Pod != &highPriNominatedPod {
		t.Errorf("Expected: %v after Pop, but got: %v", highPriNominatedPod.Name, p.Pod.Name)
	}
	if len(q.PodNominator.nominatedPods) != 1 {
		t.Errorf("Expected nomindatePods to have one element: %v", q.PodNominator)
	}

	if podInfo := q.unschedulableQ.get(&unschedulablePod); podInfo != nil && podInfo.Pod != &unschedulablePod {
		t.Errorf("Pod %v was not found in the unschedulableQ.", unschedulablePod.Name)
	}
}

// 测试AssignedPodAdded
// unschedulableQ中有个pod with affinity，这时加入一个pod with label，这时unschedulable pod 加入到 activeQ
func TestPriorityQueueAssignedPodAdded(test *testing.T) {
	affinityPod := unschedulablePod.DeepCopy()
	affinityPod.Name = "afp"
	affinityPod.Spec = v1.PodSpec{
		Affinity: &v1.Affinity{
			PodAffinity: &v1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "service",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"securityscan", "value2"},
								},
							},
						},
						TopologyKey: "region",
					},
				},
			},
		},
		Priority: &mediumPriority,
	}

	labelPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lbp",
			Namespace: affinityPod.Namespace,
			Labels:    map[string]string{"service": "securityscan"},
		},
		Spec: v1.PodSpec{NodeName: "machine1"},
	}

	c := clock.NewFakeClock(time.Now())
	q := NewPriorityQueue(newDefaultQueueSort(), WithClock(c))
	_ = q.Add(&medPriorityPod)
	_ = q.AddUnschedulableIfNotPresent(q.newQueuedPodInfo(&unschedulablePod), q.SchedulingCycle())
	_ = q.AddUnschedulableIfNotPresent(q.newQueuedPodInfo(affinityPod), q.SchedulingCycle())

	// Move clock to make the unschedulable pods complete backoff.
	c.Step(DefaultPodInitialBackoffDuration + time.Second)
	q.AssignedPodAdded(&labelPod)
	if q.unschedulableQ.get(affinityPod) != nil {
		test.Error("affinityPod is still in the unschedulableQ.")
	}
	if _, exists, _ := q.activeQ.Get(newQueuedPodInfoNoTimestamp(affinityPod)); !exists {
		test.Error("affinityPod is not moved to activeQ.")
	}
	if q.unschedulableQ.get(&unschedulablePod) == nil {
		test.Error("unschedulablePod is not in the unschedulableQ.")
	}
}
