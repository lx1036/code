package namespace

import corev1 "k8s.io/api/core/v1"

type NamespaceQuery struct {
	namespaces []string
}

func NewNamespaceQuery(namespaces []string) *NamespaceQuery {
	return &NamespaceQuery{
		namespaces: namespaces,
	}
}

func (namespaces *NamespaceQuery) GetNamespace() string  {
	if len(namespaces.namespaces) == 1 {
		return namespaces.namespaces[0]
	}
	return corev1.NamespaceAll
}
