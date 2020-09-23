package service

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/database"
	"k8s-lx1036/k8s-ui/backend/models"
	"net/http"
	"strconv"
)

type ServiceController struct {

}

func (controller *ServiceController) Delete() gin.HandlerFunc {
	return func(context *gin.Context) {
		serviceId := context.Param("serviceId")
		id, err := strconv.Atoi(serviceId)
		if err != nil {
		
		}
		
		data := database.DB.Delete(&models.Service{ID: uint(id)}).Value
		
		context.JSON(http.StatusOK, base.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   data,
		})
		
	}
}
