package types

const (
	KubernetesPodNameLabel       = "io.kubernetes.pod.name"
	KubernetesPodNamespaceLabel  = "io.kubernetes.pod.namespace"
	KubernetesPodUIDLabel        = "io.kubernetes.pod.uid"
	KubernetesContainerNameLabel = "io.kubernetes.container.name"
)

func GetContainerName(labels map[string]string) string {
	return labels[KubernetesContainerNameLabel]
}

func GetPodName(labels map[string]string) string {
	return labels[KubernetesPodNameLabel]
}

func GetPodUID(labels map[string]string) string {
	return labels[KubernetesPodUIDLabel]
}

func GetPodNamespace(labels map[string]string) string {
	return labels[KubernetesPodNamespaceLabel]
}
