package queue

import (
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/queuesort"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

func newDefaultQueueSort() framework.LessFunc {
	sort := &queuesort.PrioritySort{}
	return sort.Less
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

}
