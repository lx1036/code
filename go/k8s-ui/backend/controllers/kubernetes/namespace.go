package kubernetes

import (
	"encoding/json"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s.io/apimachinery/pkg/labels"
	"sync"
)

type KubeNamespaceController struct {
	base.APIController
}

func (controller *KubeNamespaceController) URLMapping() {
	controller.Mapping("Resources", controller.Resources)
}

// `/api/v1/kubernetes/namespaces/${namespaceId}/resources?app=${appName}`
// @router /:namespaceid([0-9]+)/resources [get]
func (controller *KubeNamespaceController) Resources() {
	appName := controller.Input().Get("app")
	namespaceId := controller.GetIntParamFromURL("namespaceId")
	namespaceModel, err := models.NamespaceModel.GetById(namespaceId)
	if err != nil {

	}
	var namespaceMetaData models.NamespaceMetaData
	err = json.Unmarshal([]byte(namespaceModel.MetaData), &namespaceMetaData)
	if err != nil {

	}

	syncResourceMap := sync.Map{}
	wg := sync.WaitGroup{}
	managers := client.Managers()
	managers.Range(func(key, value interface{}) bool {
		manager := value.(*client.ClusterManager)
		wg.Add(1)
		go func(cm *client.ClusterManager) {
			defer wg.Done()
			//clusterMetas, ok := namespaceMetaData.ClusterMetas[cm.Cluster.Name]
			//if !ok { // can't use current cluster
			//	return
			//}
			selectorMap := map[string]string{
				util.NamespaceLabelKey: namespaceModel.Name,
			}
			if appName != "" {
				selectorMap[util.AppLabelKey] = appName
			}
			selector := labels.SelectorFromSet(selectorMap)
			resourceUsage, err := namespace.ResourcesUsageByNamespace(cm.KubeClient, namespaceModel.KubeNamespace, selector.String())
			if err != nil {

			}
			syncResourceMap.Store(cm.Cluster.Name, common.Resource{
				Usage: &common.ResourceList{
					Cpu:    resourceUsage.Cpu / 1000,
					Memory: 0,
				},
				Limit: nil,
			})
		}(manager)

		return true
	})

	result := make(map[string]common.Resource)
	syncResourceMap.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(common.Resource)
		return true
	})

	controller.Success(result)
}
