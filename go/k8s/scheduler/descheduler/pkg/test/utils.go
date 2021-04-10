package test

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildTestNode creates a node with specified capacity.
func BuildTestNode(name string, millicpu int64, mem int64, pods int64, apply func(*v1.Node)) *v1.Node {
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     name,
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", name),
			Labels:   map[string]string{},
		},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourcePods:   *resource.NewQuantity(pods, resource.DecimalSI),
				v1.ResourceCPU:    *resource.NewMilliQuantity(millicpu, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(mem, resource.DecimalSI),
			},
			Allocatable: v1.ResourceList{
				v1.ResourcePods:   *resource.NewQuantity(pods, resource.DecimalSI),
				v1.ResourceCPU:    *resource.NewMilliQuantity(millicpu, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(mem, resource.DecimalSI),
			},
			Phase: v1.NodeRunning,
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}
	if apply != nil {
		apply(node)
	}
	return node
}

// SetNodeUnschedulable sets the given node unschedulable
func SetNodeUnschedulable(node *v1.Node) {
	node.Spec.Unschedulable = true
}

// BuildTestPod creates a test pod with given parameters.
func BuildTestPod(name string, cpu int64, memory int64, nodeName string, apply func(*v1.Pod)) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			SelfLink:  fmt.Sprintf("/api/v1/namespaces/default/pods/%s", name),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{},
						Limits:   v1.ResourceList{},
					},
				},
			},
			NodeName: nodeName,
		},
	}

	if cpu >= 0 {
		pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	}
	if memory >= 0 {
		pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = *resource.NewQuantity(memory, resource.DecimalSI)
	}
	if apply != nil {
		apply(pod)
	}

	return pod
}

// SetDSOwnerRef sets the given pod's owner to DaemonSet
func SetDaemonsetOwnerRef(pod *v1.Pod) {
	pod.ObjectMeta.OwnerReferences = GetDaemonSetOwnerRefList()
}

// GetDaemonSetOwnerRefList returns the ownerRef needed for daemonset pod.
func GetDaemonSetOwnerRefList() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{Kind: "DaemonSet", APIVersion: "v1"},
	}
}

// SetNormalOwnerRef sets the given pod's owner to Pod
func SetNormalOwnerRef(pod *v1.Pod) {
	pod.ObjectMeta.OwnerReferences = GetNormalPodOwnerRefList()
}

func SetPodPriority(pod *v1.Pod, priority int32) {
	pod.Spec.Priority = &priority
}

func GetNormalPodOwnerRefList() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{Kind: "Pod", APIVersion: "v1"},
	}
}

// SetRSOwnerRef sets the given pod's owner to ReplicaSet
func SetReplicaSetOwnerRef(pod *v1.Pod) {
	pod.ObjectMeta.OwnerReferences = GetReplicaSetOwnerRefList()
}

func GetReplicaSetOwnerRefList() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{Kind: "ReplicaSet", APIVersion: "v1", Name: "replicaset-1"},
	}
}

// GetMirrorPodAnnotation returns the annotation needed for mirror pod.
func GetMirrorPodAnnotation() map[string]string {
	return map[string]string{
		"kubernetes.io/created-by":    `{"kind":"SerializedReference","apiVersion":"v1","reference":{"kind":"Pod"}}`,
		"kubernetes.io/config.source": "api",
		"kubernetes.io/config.mirror": "mirror",
	}
}

func MakeBestEffortPod(pod *v1.Pod) {
	pod.Spec.Containers[0].Resources.Requests = nil
	pod.Spec.Containers[0].Resources.Limits = nil
}

func MakeBurstablePod(pod *v1.Pod) {
	pod.Spec.Containers[0].Resources.Limits = nil
}

func MakeGuaranteedPod(pod *v1.Pod) {
	pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU]
	pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory]
}
