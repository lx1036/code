package node

import (
	"context"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNodesByQuery(
	k8sClient kubernetes.Interface,
	dataSelect *dataselect.DataSelectQuery) (*NodeList, error) {
	rawNodeList, err := k8sClient.CoreV1().Nodes().List(context.TODO(), common.ListEverything)
	if err != nil {
		return nil, err
	}

	nodeList, err := toNodeList(k8sClient, rawNodeList)
	if err != nil {
		return nil, err
	}

	return nodeList, nil
}

func toNodeList(k8sClient kubernetes.Interface, rawNodeList *corev1.NodeList) (*NodeList, error) {
	nodeList := &NodeList{
		ListMeta: common.ListMeta{TotalItems: len(rawNodeList.Items)},
	}

	for _, rawNode := range rawNodeList.Items {
		podList := getNodePods(k8sClient, rawNode)
		nodeList.Nodes = append(nodeList.Nodes, toNode(rawNode, podList))
	}

	return nodeList, nil
}

func toNode(rawNode corev1.Node, podList *corev1.PodList) Node {
	ready := getNodeConditionStatus(rawNode, corev1.NodeReady)
	allocatedResources, _ := getNodeAllocatedResources(rawNode, podList)

	return Node{
		ObjectMeta:         common.NewObjectMeta(rawNode.ObjectMeta),
		TypeMeta:           common.NewTypeMeta(common.ResourceKindNode),
		Ready:              ready,
		AllocatedResources: allocatedResources,
	}
}

func getNodePods(k8sClient kubernetes.Interface, rawNode corev1.Node) *corev1.PodList {

}

// ???
func getNodeConditionStatus(node corev1.Node, conditionType corev1.NodeConditionType) corev1.ConditionStatus {
	for _, condition := range node.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status
		}
	}
	return corev1.ConditionUnknown
}

// 获取该 Node 上可用资源 cpu/memory
func getNodeAllocatedResources(rawNode corev1.Node, podList *corev1.PodList) (NodeAllocatedResources, error) {

	return NodeAllocatedResources{
		CPURequests:            0,
		CPURequestsFraction:    0,
		CPULimits:              0,
		CPULimitsFraction:      0,
		CPUCapacity:            0,
		MemoryRequests:         0,
		MemoryRequestsFraction: 0,
		MemoryLimits:           0,
		MemoryLimitsFraction:   0,
		MemoryCapacity:         0,
		AllocatedPods:          0,
		PodCapacity:            0,
		PodFraction:            0,
	}, nil
}
