package kubernetes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/client"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/resources/pod"
	"net/http"
	"sync"
)

type KubePodController struct {
}

//func (controller *KubePodController) URLMapping() {
//	controller.Mapping("PodStatistics", controller.PodStatistics)
//
//}



func (controller *KubePodController) PodStatistics() gin.HandlerFunc {
	return func(context *gin.Context) {
		type PodStatistics struct {
			Total   int            `json:"total"`
			Details map[string]int `json:"details"`
		}
		total := 0
		countMap := make(map[string]int)
		countSyncMap := sync.Map{}
		cluster := context.Query("cluster")
		if cluster == "" {
			managers := client.Managers()
			wg := sync.WaitGroup{}

			managers.Range(func(key, value interface{}) bool {
				manager := value.(*client.ClusterManager)
				wg.Add(1)
				go func(manager *client.ClusterManager) {
					defer wg.Done()
					count, err := pod.GetPodCounts(manager.CacheFactory)
					if err != nil {
						context.JSON(http.StatusInternalServerError, base.JsonResponse{
							Errno:  -1,
							Errmsg: fmt.Sprintf("failed: get pod counts [%s]", err.Error()),
							Data:   nil,
						})
						return
					}
					total += count
					countSyncMap.Store(manager.Cluster.Name, count)
				}(manager)

				return true
			})

			wg.Wait()
		} else {

		}

		context.JSON(http.StatusOK, base.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data: PodStatistics{
				Total:   total,
				Details: countMap,
			},
		})
	}
}
