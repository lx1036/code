package namespace

import (
	"context"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/limitrange"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/resourcequota"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func toNamespace(rawNamespace corev1.Namespace) Namespace {
	return Namespace{
		ObjectMeta: common.NewObjectMeta(rawNamespace.ObjectMeta),
		TypeMeta:   common.NewTypeMeta(common.ResourceKindNamespace),
		Phase:      rawNamespace.Status.Phase,
	}
}

// TODO: refactor to use channel to concurrent-io
func GetNamespaceByQuery(
	k8sClient kubernetes.Interface,
	dataSelect *dataselect.DataSelectQuery,
	namespaceName string) (*NamespaceDetail, error) {
	rawNamespace, err := k8sClient.CoreV1().Namespaces().Get(context.TODO(), namespaceName, common.GetEverything)
	if err != nil {
		return nil, err
	}
	
	resourceQuotaDetailList, err := getResourceQuotas(k8sClient, *rawNamespace)
	if err != nil {
		return nil, err
	}
	
	limitRangeItems, err := getLimitRanges(k8sClient, *rawNamespace)
	if err != nil {
		return nil, err
	}
	
	return &NamespaceDetail{
		Namespace: toNamespace(*rawNamespace),
		ResourceLimits: limitRangeItems,
		ResourceQuotaList: resourceQuotaDetailList,
	}, nil
}

func CreateNamespaceByQuery(k8sClient kubernetes.Interface, namespaceName string) (Namespace, error) {
	rawNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	namespace, err := k8sClient.CoreV1().Namespaces().Create(context.TODO(), rawNamespace, common.CreateEverything)
	if err != nil {
		return Namespace{}, err
	}
	
	return toNamespace(*namespace), nil
}

func getResourceQuotas(k8sClient kubernetes.Interface, rawNamespace corev1.Namespace) (*resourcequota.ResourceQuotaDetailList, error) {
	resourceQuotasList, err :=  k8sClient.CoreV1().ResourceQuotas(rawNamespace.Name).List(context.TODO(), common.ListEverything)
	if err != nil {
		return nil, err
	}
	result := &resourcequota.ResourceQuotaDetailList{
		ListMeta: common.ListMeta{
			TotalItems: len(resourceQuotasList.Items),
		},
	}
	
	for _, item := range resourceQuotasList.Items {
		result.Items = append(result.Items, resourcequota.ToResourceQuotaDetail(&item))
	}
	
	return result, nil
}

func getLimitRanges(k8sClient kubernetes.Interface, rawNamespace corev1.Namespace) ([]limitrange.LimitRangeItem, error) {
	limitRange, err := k8sClient.CoreV1().LimitRanges(rawNamespace.Name).List(context.TODO(), common.ListEverything)
	if err != nil {
		return nil, err
	}
	var limitRanges []limitrange.LimitRangeItem
	for _, item := range limitRange.Items {
		limitRanges = append(limitRanges, limitrange.ToLimitRangeItem(&item)...)
	}
	
	return limitRanges, nil
}


