package pod

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/client"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"net/http"
	"strings"
)

type PodController struct {
}

func (controller *PodController) ListNamespacePod() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()

		namespace := parseNamespace(context.Param("namespace"))
		common.ParseDataSelectFromRequest(context)
		result := ListPod(k8sClient)

		return context.JSON(http.StatusOK, gin.H{

		})
	}
}

func parseNamespace(namespace string) *common.NamespaceQuery {
	namespaces := strings.Split(namespace, ",")
	var noEmptyNamespaces []string
	for _, value := range namespaces {
		if len(strings.TrimSpace(value)) != 0 {
			noEmptyNamespaces = append(noEmptyNamespaces, value)
		}
	}

	return common.NewNamespaceQuery(noEmptyNamespaces)
}
