package controller

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
)

const (
	randomSuffixLength = 10
)

func GetPodNames(pods []*corev1.Pod) []string {
	if len(pods) == 0 {
		return nil
	}
	res := []string{}
	for _, p := range pods {
		res = append(res, p.Name)
	}
	return res
}

func GetLabelsForEtcdPod(etcdClusterName string) map[string]string {
	return map[string]string{
		"app":          "etcd",
		"etcd_cluster": etcdClusterName,
	}
}

func LabelsForCluster(clusterName string) map[string]string {
	return map[string]string{
		"etcd_cluster": clusterName,
		"app":          "etcd",
	}
}

func UniqueMemberName(name string) string {
	return fmt.Sprintf("%s-%s", name, utilrand.String(randomSuffixLength))
}
