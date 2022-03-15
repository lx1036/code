package controller

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/kubernetes/pkg/apis/networking"
)

func (controller *NetworkPolicyController) ListPodsByNamespaceAndLabels(namespace string, podSelector labels.Selector) (ret []*corev1.Pod, err error) {
	return listerv1.NewPodLister(controller.podLister).Pods(namespace).List(podSelector)
}

func isNetworkPolicyPod(pod *corev1.Pod) bool {
	return len(pod.Status.PodIP) != 0 && !pod.Spec.HostNetwork && pod.Status.Phase == corev1.PodRunning
}

func listPodIPBlock(peer networking.NetworkPolicyPeer) [][]string {
	ipBlock := make([][]string, 0)
	if peer.PodSelector == nil && peer.NamespaceSelector == nil && peer.IPBlock != nil {
		ipBlock = append(ipBlock, []string{peer.IPBlock.CIDR, "timeout", "0"})
		for _, except := range peer.IPBlock.Except {
			ipBlock = append(ipBlock, []string{except, "timeout", "0", "nomatch"})
		}
	}

	return ipBlock
}
