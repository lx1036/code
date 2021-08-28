package controller

import corev1 "k8s.io/api/core/v1"

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
