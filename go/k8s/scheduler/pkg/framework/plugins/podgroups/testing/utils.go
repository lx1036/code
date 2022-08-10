package testing

import (
	"fmt"
	schedulertesting "k8s-lx1036/k8s/scheduler/pkg/testing"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	podgroupv1 "k8s-lx1036/k8s/scheduler/pkg/apis/podgroup/v1"
)

func MakePodGroup(name, namespace string, min int32, creationTime *time.Time, minResource *corev1.ResourceList) *podgroupv1.PodGroup {
	var ti int32 = 10
	pg := &podgroupv1.PodGroup{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       podgroupv1.PodGroupSpec{MinMember: min, ScheduleTimeoutSeconds: &ti},
	}
	if creationTime != nil {
		pg.CreationTimestamp = metav1.Time{Time: *creationTime}
	}
	if minResource != nil {
		pg.Spec.MinResources = minResource
	}
	return pg
}

func MakeNodesAndPods(labelsForPod map[string]string, existingPodsNum, allNodesNum int) (existingPods []*corev1.Pod, allNodes []*corev1.Node) {
	type keyVal struct {
		k string
		v string
	}
	var labelPairs []keyVal
	for k, v := range labelsForPod {
		labelPairs = append(labelPairs, keyVal{k: k, v: v})
	}
	// build nodes
	for i := 0; i < allNodesNum; i++ {
		res := map[corev1.ResourceName]string{
			corev1.ResourceCPU:  "1",
			corev1.ResourcePods: "20",
		}
		node := schedulertesting.MakeNode().Name(fmt.Sprintf("node%d", i)).Capacity(res)
		allNodes = append(allNodes, &node.Node)
	}
	// build pods
	for i := 0; i < existingPodsNum; i++ {
		podWrapper := schedulertesting.MakePod().Name(fmt.Sprintf("pod%d", i)).Node(fmt.Sprintf("node%d", i%allNodesNum))
		// apply labels[0], labels[0,1], ..., labels[all] to each pod in turn
		for _, p := range labelPairs[:i%len(labelPairs)+1] {
			podWrapper = podWrapper.Label(p.k, p.v)
		}
		existingPods = append(existingPods, podWrapper.Obj())
	}
	return
}
