package pod

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/client"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/namespace"
	"net/http"
	"strings"
)

type PodController struct {
}

func (controller *PodController) ListPods() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceQuery := parseNamespace(context.Param("namespace"))
		dataselectQuery := dataselect.ParseDataSelectFromRequest(context)
		result := ListPod(k8sClient, namespaceQuery, dataselectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   result,
		})
	}
}

func (controller *PodController) GetPod() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func (controller *PodController) ListPodContainers() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func (controller *PodController) ExecPodShell() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func (controller *PodController) GetPodPvc() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func parseNamespace(namespaceQuery string) *namespace.NamespaceQuery {
	namespaces := strings.Split(namespaceQuery, ",")
	var noEmptyNamespaces []string
	for _, value := range namespaces {
		if len(strings.TrimSpace(value)) != 0 {
			noEmptyNamespaces = append(noEmptyNamespaces, value)
		}
	}

	return namespace.NewNamespaceQuery(noEmptyNamespaces)
}
