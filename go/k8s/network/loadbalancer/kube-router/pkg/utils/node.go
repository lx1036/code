package utils

import (
	"context"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net"
	"os"

	corev1 "k8s.io/api/core/v1"
)

// GetNodeIP returns the most valid external facing IP address for a node.
// Order of preference:
// 1. NodeInternalIP
// 2. NodeExternalIP (Only set on cloud providers usually)
func GetNodeIP(node *corev1.Node) (net.IP, error) {
	addresses := node.Status.Addresses
	addressMap := make(map[corev1.NodeAddressType][]corev1.NodeAddress)
	for i := range addresses {
		addressMap[addresses[i].Type] = append(addressMap[addresses[i].Type], addresses[i])
	}
	if addresses, ok := addressMap[corev1.NodeInternalIP]; ok {
		return net.ParseIP(addresses[0].Address), nil
	}
	if addresses, ok := addressMap[corev1.NodeExternalIP]; ok {
		return net.ParseIP(addresses[0].Address), nil
	}

	return nil, errors.New("host IP unknown")
}

const (
	PodCIDRAnnotation = "kube-router.io/pod-cidr"
)

func GetPodCidrFromNodeSpec(clientset kubernetes.Interface) (string, error) {
	node, err := GetNodeObject(clientset)
	if err != nil {
		return "", fmt.Errorf("Failed to get pod CIDR allocated for the node due to: " + err.Error())
	}

	if cidr, ok := node.Annotations[PodCIDRAnnotation]; ok {
		_, _, err = net.ParseCIDR(cidr)
		if err != nil {
			return "", fmt.Errorf("error parsing pod CIDR in node annotation: %v", err)
		}

		return cidr, nil
	}

	if node.Spec.PodCIDR == "" {
		return "", fmt.Errorf("node.Spec.PodCIDR not set for node: %v", node.Name)
	}

	return node.Spec.PodCIDR, nil
}

func GetNodeObject(clientset kubernetes.Interface) (*corev1.Node, error) {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName != "" {
		node, err := clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
		if err == nil {
			return node, nil
		}
	}

	// if env NODE_NAME is not set then check if node is register with hostname
	hostName, _ := os.Hostname()
	node, err := clientset.CoreV1().Nodes().Get(context.Background(), hostName, metav1.GetOptions{})
	if err == nil {
		return node, nil
	}

	return nil, fmt.Errorf("failed to identify the node by NODE_NAME, hostname or --hostname-override")
}

const (
	IPInIPHeaderLength = 20
)
