package deployment

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/dashboard/client"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/http"
)

type DeploymentController struct {
}
type PortMapping struct {
	// Port that will be exposed on the service.
	Port int32 `json:"port"`
	// Docker image path for the application.
	TargetPort int32 `json:"targetPort"`
	// IP protocol for the mapping, e.g., "TCP" or "UDP".
	Protocol api.Protocol `json:"protocol"`
}
type DeploymentSpec struct {
	Name            string            `json:"name" form:"name" binding:"required"`
	Namespace       string            `json:"namespace" form:"namespace" binding:"required"`
	Labels          []Label           `json:"labels" form:"labels" binding:"required"`
	Description     string            `json:"description" form:"description"`
	Replicas        int32             `json:"replicas" form:"replicas" binding:"required"`
	ContainerImage  string            `json:"containerImage" form:"containerImage" binding:"required"`
	RunAsPrivileged bool              `json:"runAsPrivileged" form:"runAsPrivileged"`
	PortMappings    []PortMapping     `json:"portMappings" form:"portMappings"`
	IsExternal      bool              `json:"isExternal" form:"isExternal"`
	RestartPolicy   api.RestartPolicy `json:"restartPolicy" form:"restartPolicy"`
}

type Label struct {
	// Label key
	Key string `json:"key"`
	// Label value
	Value string `json:"value"`
}

const (
	DescriptionAnnotationKey = "description"
)

func (controller *DeploymentController) HandleDeploy() gin.HandlerFunc {
	return func(context *gin.Context) {
		var spec DeploymentSpec
		if err := context.ShouldBindJSON(&spec); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"errno":  -1,
				"errmsg": err.Error(),
				"data":   nil,
			})
			return
		}

		k8sClient := client.DefaultClientManager.Client()
		annotations := map[string]string{}
		if spec.Description != "" {
			annotations[DescriptionAnnotationKey] = spec.Description
		}
		labels := getLabelsMap(spec.Labels)
		objectMeta := metaV1.ObjectMeta{
			Name:        spec.Name,
			Labels:      labels,
			Annotations: annotations,
		}
		containerSpec := api.Container{
			Name:  spec.Name,
			Image: spec.ContainerImage,
			Resources: api.ResourceRequirements{
				Requests: make(map[api.ResourceName]resource.Quantity),
			},
			SecurityContext: &api.SecurityContext{
				Privileged: &spec.RunAsPrivileged,
			},
		}

		restartPolicy := spec.RestartPolicy
		if restartPolicy == "" {
			restartPolicy = api.RestartPolicyAlways // can't be api.RestartPolicyNever, why?
		}

		deployment := &apps.Deployment{
			ObjectMeta: objectMeta,
			Spec: apps.DeploymentSpec{
				Replicas: &spec.Replicas,
				Selector: &metaV1.LabelSelector{
					MatchLabels: labels,
				},
				Template: api.PodTemplateSpec{
					ObjectMeta: objectMeta,
					Spec: api.PodSpec{
						Containers:    []api.Container{containerSpec},
						RestartPolicy: restartPolicy,
					},
				},
			},
		}

		var deploymentNew *v1.Deployment
		var err1 error
		_, err := k8sClient.AppsV1().Deployments(spec.Namespace).Get(deployment.Name, metaV1.GetOptions{})
		if err != nil {
			deploymentNew, err1 = k8sClient.AppsV1().Deployments(spec.Namespace).Create(deployment) // create
		} else {
			deploymentNew, err1 = k8sClient.AppsV1().Deployments(spec.Namespace).Update(deployment) // update
		}

		if err1 != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"errno":  -1,
				"errmsg": err1.Error(),
				"data":   nil,
			})
			return
		}

		var serviceNew *api.Service
		if len(spec.PortMappings) > 0 {
			service := &api.Service{
				ObjectMeta: objectMeta,
				Spec: api.ServiceSpec{
					Selector: labels,
				},
			}

			if spec.IsExternal {
				service.Spec.Type = api.ServiceTypeNodePort
			} else {
				service.Spec.Type = api.ServiceTypeClusterIP
			}

			for _, portMapping := range spec.PortMappings {
				servicePort := api.ServicePort{
					Name:     spec.Name,
					Protocol: portMapping.Protocol,
					Port:     portMapping.Port,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: portMapping.TargetPort,
					},
				}
				service.Spec.Ports = append(service.Spec.Ports, servicePort)
			}

			var err1 error
			_, err := k8sClient.CoreV1().Services(spec.Namespace).Get(service.Name, metaV1.GetOptions{})
			if err != nil {
				serviceNew, err1 = k8sClient.CoreV1().Services(spec.Namespace).Create(service) // create
			} else {
				service.ObjectMeta.SetResourceVersion("1.0.1")
				serviceNew, err1 = k8sClient.CoreV1().Services(spec.Namespace).Update(service) // update
			}
			if err1 != nil {
				context.JSON(http.StatusInternalServerError, gin.H{
					"errno":  -1,
					"errmsg": err1.Error(),
					"data":   nil,
				})
				return
			}
		}

		context.JSON(http.StatusOK, gin.H{
			"errno":  0,
			"errmsg": "success",
			"data": struct {
				Deployment *v1.Deployment `json:"deployment"`
				Service    *api.Service   `json:"service"`
			}{
				Deployment: deploymentNew,
				Service:    serviceNew,
			},
		})
	}
}

type AppNameValiditySpec struct {
	Name      string `json:"name" form:"name" binding:"required"`
	Namespace string `json:"namespace" form:"namespace" binding:"required"`
}

func (controller *DeploymentController) HandleNameValidity() gin.HandlerFunc {
	return func(context *gin.Context) {
		var spec AppNameValiditySpec
		if err := context.ShouldBindJSON(&spec); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"errno":  -1,
				"errmsg": err.Error(),
				"data":   nil,
			})
			return
		}

		isDeploymentValid := false
		isServiceValid := false
		k8sClient := client.DefaultClientManager.Client()
		_, err := k8sClient.AppsV1().Deployments(spec.Namespace).Get(spec.Name, metaV1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) || errors.IsForbidden(err) {
				isDeploymentValid = true
			}
		} else {
			isDeploymentValid = true
		}

		_, err = k8sClient.CoreV1().Services(spec.Namespace).Get(spec.Name, metaV1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) || errors.IsForbidden(err) {
				isServiceValid = true
			}
		} else {
			isServiceValid = true
		}

		isValid := isDeploymentValid && isServiceValid
		if isValid {
			context.JSON(http.StatusOK, gin.H{
				"errno":  0,
				"errmsg": "success",
				"data": struct {
					Valid bool `json:"valid"`
				}{
					Valid: isValid,
				},
			})
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{
				"errno":  0,
				"errmsg": "fail",
				"data": struct {
					Valid bool `json:"valid"`
				}{
					Valid: isValid,
				},
			})
		}
	}
}

// Converts array of labels to map[string]string
func getLabelsMap(labels []Label) map[string]string {
	result := make(map[string]string)
	for _, label := range labels {
		result[label.Key] = label.Value
	}
	return result
}
