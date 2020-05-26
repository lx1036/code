package pod

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/client"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/namespace"
	"net/http"
	"strings"
)

type PodController struct {
}

func (controller *PodController) ListNamespacePod() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()

		namespaceQuery := parseNamespace(context.Param("namespace"))
		dataselectQuery := dataselect.ParseDataSelectFromRequest(context)
		result := ListPod(k8sClient, namespaceQuery, dataselectQuery)

		return context.JSON(http.StatusOK, gin.H{

		})
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
