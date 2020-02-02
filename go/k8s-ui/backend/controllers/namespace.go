package controllers

import (
	"k8s-lx1036/k8s-ui/backend/common"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
	"strconv"
)

type NamespaceController struct {
	base.APIController
}

func (controller *NamespaceController) URLMapping() {
	controller.Mapping("Statistics", controller.Statistics)
}

func (controller *NamespaceController) Statistics() {
	namespaceId := controller.GetIntParamFromURL("namespaceId")
	appId, err := strconv.ParseInt(controller.Input().Get("appId"), 10, 64)
	if err != nil {

	}
	param := &common.QueryParam{
		Query: map[string]interface{}{
			"deleted_at":         "",
			"App__Namespace__Id": namespaceId,
		},
	}

	if appId != 0 {
		param.Query["App__Id"] = appId
	}

	resources := []string{
		models.TableNameDeployment,
		models.TableNameStatefulset,
		models.TableNameDaemonSet,
		models.TableNameCronjob,
		models.TableNameService,
		models.TableNameConfigMap,
		models.TableNameSecret,
		models.TableNamePersistentVolumeClaim,
	}

	var results = make(map[string]int64)

	for _, resource := range resources {
		count, err := models.GetTotalCount(resource, param)
		if err != nil {

		}
		kubeApiType := models.TableToKubeApiTypeMap[resource]
		results[string(kubeApiType)] = count
	}

	controller.Success(results)
}
