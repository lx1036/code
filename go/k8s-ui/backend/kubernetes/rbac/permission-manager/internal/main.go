package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	kubeClient kubernetes.Interface
)

func main() {
	viper.AutomaticEnv()

	var err error
	kubeClient, err = GetKubeClient()
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	api := router.Group("api/v1")
	api.Use(Auth())
	{
		api.GET("/users", ListUsers())
		api.GET("/namespaces", ListNamespaces())
		api.GET("/rbac", ListRbac())

		api.POST("/users", CreateUser())
		api.POST("/cluster-role", CreateClusterRole())
		api.POST("/cluster-role-binding", CreateClusterRoleBinding())
		api.POST("/role", CreateRole())
		api.POST("/role-binding", CreateRoleBinding())

		api.POST("/delete-user", DeleteUser())
		api.POST("/delete-cluster-role", DeleteClusterRole())
		api.POST("/delete-cluster-role-binding", DeleteClusterRoleBinding())
		api.POST("/delete-role", DeleteRole())
		api.POST("/delete-role-binding", DeleteRoleBinding())

		api.POST("/create-kubeconfig", CreateKubeConfig())
	}

	fmt.Println(router.Run())
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

const (
	ResourceUserUrl = "/apis/permissionmanager.user/v1alpha1/permissionmanagerusers"
)

type PermissionManagerUserMetadata struct {
	Annotations struct {
		KubectlKubernetesIoLastAppliedConfiguration string `json:"kubectl.kubernetes.io/last-applied-configuration"`
	} `json:"annotations"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
	Generation        int       `json:"generation"`
	Name              string    `json:"name"`
	ResourceVersion   string    `json:"resourceVersion"`
	SelfLink          string    `json:"selfLink"`
	UID               string    `json:"uid"`
}
type PermissionManagerUserSpec struct {
	Name string `json:"name"`
}
type PermissionManagerUser struct {
	APIVersion string                        `json:"apiVersion"`
	Kind       string                        `json:"kind"`
	Metadata   PermissionManagerUserMetadata `json:"metadata"`
	Spec       PermissionManagerUserSpec     `json:"spec"`
}

type PermissionManagerUsers struct {
	APIVersion string                  `json:"apiVersion"`
	Items      []PermissionManagerUser `json:"items"`
	Kind       string                  `json:"kind"`
	Metadata   struct {
		Continue        string `json:"continue"`
		ResourceVersion string `json:"resourceVersion"`
		SelfLink        string `json:"selfLink"`
	} `json:"metadata"`
}

func ListUsers() gin.HandlerFunc {
	return func(context *gin.Context) {
		raw, err := kubeClient.AppsV1().RESTClient().Get().AbsPath(ResourceUserUrl).DoRaw()
		var permissionManagerUsers PermissionManagerUsers
		err = json.Unmarshal(raw, &permissionManagerUsers)
		if err != nil {
			context.JSON(http.StatusBadRequest, Response{
				Code:    -1,
				Message: "bad body",
			})
			return
		}

		var users []PermissionManagerUserSpec
		for _, item := range permissionManagerUsers.Items {
			users = append(users, PermissionManagerUserSpec{
				Name: item.Spec.Name,
			})
		}
		context.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "success",
			Data:    users,
		})
	}
}

func ListNamespaces() gin.HandlerFunc {
	return func(context *gin.Context) {
		namespaces, err := kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			context.JSON(http.StatusBadRequest, Response{
				Code:    -1,
				Message: "bad body",
			})
			return
		}

		var names []string
		for _, namespace := range namespaces.Items {
			names = append(names, namespace.Name)
		}

		context.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "success",
			Data:    names,
		})
	}
}

func ListRbac() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func CreateUser() gin.HandlerFunc {
	return func(context *gin.Context) {
		var body struct {
			Username string `json:"username" binding:"required"`
		}
		_ = context.BindJSON(&body)
		metadataName := "permissionmanager.user." + body.Username
		var payload = PermissionManagerUser{
			APIVersion: "permissionmanager.user/v1alpha1",
			Kind:       "Permissionmanageruser",
			Metadata: PermissionManagerUserMetadata{
				Name: metadataName,
			},
			Spec: PermissionManagerUserSpec{
				Name: body.Username,
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		var err error
		raw, err := kubeClient.AppsV1().RESTClient().Post().AbsPath(ResourceUserUrl).Body(payloadBytes).DoRaw()
		var permissionManagerUser PermissionManagerUser
		err = json.Unmarshal(raw, &permissionManagerUser)
		if err != nil {
			context.JSON(http.StatusBadRequest, Response{
				Code:    -1,
				Message: "bad body",
			})
			return
		}

		context.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "success",
			Data:    permissionManagerUser,
		})
	}
}

func CreateClusterRole() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func CreateClusterRoleBinding() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func CreateRole() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func CreateRoleBinding() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func DeleteUser() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func DeleteClusterRole() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func DeleteClusterRoleBinding() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func DeleteRole() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func DeleteRoleBinding() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func CreateKubeConfig() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func GetKubeConfig() (*restclient.Config, error) {
	var kubeconfig *string
	if home, _ := os.UserHomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetKubeClient() (kubernetes.Interface, error) {
	var config *restclient.Config
	var err error
	if len(viper.GetString("KUBERNETES_HOST")) != 0 { // inside k8s
		config, err = restclient.InClusterConfig()
	} else {
		config, err = GetKubeConfig()
	}

	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}
