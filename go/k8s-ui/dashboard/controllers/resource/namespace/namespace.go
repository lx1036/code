package namespace

import (
	"context"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

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

func ListNamespacesByQuery(
	k8sClient kubernetes.Interface,
	dataSelect *dataselect.DataSelectQuery) (*NamespaceList, error) {
	rawNamespaces, err := k8sClient.CoreV1().Namespaces().List(context.TODO(), common.ListEverything)
	if err != nil {
		return nil, err
	}
	
	namespaceList := &NamespaceList{
		ListMeta: common.ListMeta{
			TotalItems: len(rawNamespaces.Items),
		},
	}
	var namespaces []Namespace
	for _, namespace := range rawNamespaces.Items {
		namespaces = append(namespaces, toNamespace(namespace))
	}
	namespaceList.Namespaces = namespaces
	
	return namespaceList, nil
}

func toNamespace(namespace corev1.Namespace) Namespace {
	return Namespace{
		ObjectMeta: common.NewObjectMeta(namespace.ObjectMeta),
		TypeMeta:   common.NewTypeMeta(common.ResourceKindNamespace),
		Phase:      namespace.Status.Phase,
	}
}
