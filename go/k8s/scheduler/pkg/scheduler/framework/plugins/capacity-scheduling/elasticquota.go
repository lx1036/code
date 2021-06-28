package capacity_scheduling

import (
	//framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type ElasticQuotaInfos map[string]*ElasticQuotaInfo

func NewElasticQuotaInfos() ElasticQuotaInfos {
	return make(ElasticQuotaInfos)
}

// ElasticQuotaInfo is a wrapper to a ElasticQuota with information.
// Each namespace can only have one ElasticQuota.
type ElasticQuotaInfo struct {
	Namespace string
	pods      sets.String
	Min       *framework.Resource
	Max       *framework.Resource
	Used      *framework.Resource
}

func newElasticQuotaInfo(namespace string, min, max, used v1.ResourceList) *ElasticQuotaInfo {
	elasticQuotaInfo := &ElasticQuotaInfo{
		Namespace: namespace,
		pods:      sets.NewString(),
		Min:       framework.NewResource(min),
		Max:       framework.NewResource(max),
		Used:      framework.NewResource(used),
	}

	return elasticQuotaInfo
}

// INFO: 更新 pods/Used
func (e *ElasticQuotaInfo) addPodIfNotPresent(pod *v1.Pod) error {
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return err
	}

	if e.pods.Has(key) {
		return nil
	}

	e.pods.Insert(key)

	podRequest := computePodResourceRequest(pod)
	e.reserveResource(podRequest.Resource)

	return nil
}

func (e *ElasticQuotaInfo) reserveResource(request framework.Resource) {
	e.Used.Memory += request.Memory
	e.Used.MilliCPU += request.MilliCPU

	for name, value := range request.ScalarResources {
		e.Used.SetScalar(name, e.Used.ScalarResources[name]+value)
	}
}

func (e *ElasticQuotaInfo) deletePodIfPresent(pod *v1.Pod) error {
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return err
	}

	if !e.pods.Has(key) {
		return nil
	}

	e.pods.Delete(key)
	podRequest := computePodResourceRequest(pod)
	e.unreserveResource(podRequest.Resource)

	return nil
}

func (e *ElasticQuotaInfo) unreserveResource(request framework.Resource) {
	e.Used.Memory -= request.Memory
	e.Used.MilliCPU -= request.MilliCPU
	for name, value := range request.ScalarResources {
		e.Used.SetScalar(name, e.Used.ScalarResources[name]-value)
	}
}
