package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s/storage/etcd/ui/backend/common"
	"net/http"
)

type EtcdController struct {
}

func (controller *EtcdController) ListMembers() gin.HandlerFunc {
	return func(context *gin.Context) {
		etcdName, ok := context.Get("etcd_name")
		if !ok {
			context.JSON(http.StatusBadRequest, common.JsonResponse{
				Errno:  -1,
				Errmsg: "etcd_server is needed",
				Data:   nil,
			})
		}
		
		fmt.Println(etcdName)
	}
}

func (controller *EtcdController) ListServers() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}

func (controller *EtcdController) CreateKey() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}

func (controller *EtcdController) List() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}

func (controller *EtcdController) GetKey() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}

func (controller *EtcdController) UpdateKey() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}

func (controller *EtcdController) DeleteKey() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}

func (controller *EtcdController) GetKeyFormat() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}

func (controller *EtcdController) GetLogs() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}

func (controller *EtcdController) GetUsers() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}

func (controller *EtcdController) GetLogTypes() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}
