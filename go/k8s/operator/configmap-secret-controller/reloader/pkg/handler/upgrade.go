package handler

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/cmd/options"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/kube"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/metrics"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"strings"
)

// Result is a status for deployment update
type Result int

const (
	// Updated is returned when environment variable is created/updated
	Updated Result = 1 + iota
	// NotUpdated is returned when environment variable is found but had value equals to the new value
	NotUpdated
	// NoEnvVarFound is returned when no environment variable is found
	NoEnvVarFound
	// NoContainerFound is returned when no environment variable is found
	NoContainerFound
)

const (
	// EnvVarPrefix is a Prefix for environment variable
	EnvVarPrefix = "STAKATER_"
)

//ItemsFunc is a generic function to return a specific resource array in given namespace
type ItemsFunc func(*kubernetes.Clientset, string) []interface{}

//ContainersFunc is a generic func to return containers
type ContainersFunc func(interface{}) []corev1.Container

//InitContainersFunc is a generic func to return containers
type InitContainersFunc func(interface{}) []corev1.Container

//VolumesFunc is a generic func to return volumes
type VolumesFunc func(interface{}) []corev1.Volume

//UpdateFunc performs the resource update
type UpdateFunc func(*kubernetes.Clientset, string, interface{}) error

//AnnotationsFunc is a generic func to return annotations
type AnnotationsFunc func(interface{}) map[string]string

//PodAnnotationsFunc is a generic func to return annotations
type PodAnnotationsFunc func(interface{}) map[string]string

//RollingUpgradeFuncs contains generic functions to perform rolling upgrade
type RollingUpgradeFuncs struct {
	ItemsFunc          ItemsFunc
	AnnotationsFunc    AnnotationsFunc
	PodAnnotationsFunc PodAnnotationsFunc
	ContainersFunc     ContainersFunc
	InitContainersFunc InitContainersFunc
	UpdateFunc         UpdateFunc
	VolumesFunc        VolumesFunc
	ResourceType       string
}

// GetDeploymentRollingUpgradeFuncs returns all callback funcs for a deployment
func GetDeploymentRollingUpgradeFuncs() RollingUpgradeFuncs {
	return RollingUpgradeFuncs{
		ItemsFunc:          GetDeploymentItems,
		AnnotationsFunc:    GetDeploymentAnnotations,
		PodAnnotationsFunc: GetDeploymentPodAnnotations,
		ContainersFunc:     GetDeploymentContainers,
		InitContainersFunc: GetDeploymentInitContainers,
		UpdateFunc:         UpdateDeployment,
		VolumesFunc:        GetDeploymentVolumes,
		ResourceType:       "Deployment",
	}
}

// GetDaemonSetRollingUpgradeFuncs returns all callback funcs for a daemonset
func GetDaemonSetRollingUpgradeFuncs() RollingUpgradeFuncs {
	return RollingUpgradeFuncs{
		ItemsFunc:          GetDaemonSetItems,
		AnnotationsFunc:    GetDaemonSetAnnotations,
		PodAnnotationsFunc: GetDaemonSetPodAnnotations,
		ContainersFunc:     GetDaemonSetContainers,
		InitContainersFunc: GetDaemonSetInitContainers,
		UpdateFunc:         UpdateDaemonSet,
		VolumesFunc:        GetDaemonSetVolumes,
		ResourceType:       "DaemonSet",
	}
}

// GetStatefulSetRollingUpgradeFuncs returns all callback funcs for a statefulSet
func GetStatefulSetRollingUpgradeFuncs() RollingUpgradeFuncs {
	return RollingUpgradeFuncs{
		ItemsFunc:          GetStatefulSetItems,
		AnnotationsFunc:    GetStatefulSetAnnotations,
		PodAnnotationsFunc: GetStatefulSetPodAnnotations,
		ContainersFunc:     GetStatefulSetContainers,
		InitContainersFunc: GetStatefulSetInitContainers,
		UpdateFunc:         UpdateStatefulSet,
		VolumesFunc:        GetStatefulSetVolumes,
		ResourceType:       "StatefulSet",
	}
}

// GetDeploymentItems returns the deployments in given namespace
func GetDeploymentItems(client *kubernetes.Clientset, namespace string) []interface{} {
	deployments, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Failed to list deployments %v", err)
	}
	return util.InterfaceSlice(deployments.Items)
}

// GetDaemonSetItems returns the daemonSets in given namespace
func GetDaemonSetItems(client *kubernetes.Clientset, namespace string) []interface{} {
	daemonSets, err := client.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Failed to list daemonSets %v", err)
	}
	return util.InterfaceSlice(daemonSets.Items)
}

// GetStatefulSetItems returns the statefulSets in given namespace
func GetStatefulSetItems(client *kubernetes.Clientset, namespace string) []interface{} {
	statefulSets, err := client.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Failed to list statefulSets %v", err)
	}
	return util.InterfaceSlice(statefulSets.Items)
}

// GetDeploymentAnnotations returns the annotations of given deployment
func GetDeploymentAnnotations(item interface{}) map[string]string {
	return item.(appsv1.Deployment).ObjectMeta.Annotations
}

// GetDaemonSetAnnotations returns the annotations of given daemonSet
func GetDaemonSetAnnotations(item interface{}) map[string]string {
	return item.(appsv1.DaemonSet).ObjectMeta.Annotations
}

// GetStatefulSetAnnotations returns the annotations of given statefulSet
func GetStatefulSetAnnotations(item interface{}) map[string]string {
	return item.(appsv1.StatefulSet).ObjectMeta.Annotations
}

// GetDeploymentPodAnnotations returns the pod's annotations of given deployment
func GetDeploymentPodAnnotations(item interface{}) map[string]string {
	return item.(appsv1.Deployment).Spec.Template.ObjectMeta.Annotations
}

// GetDaemonSetPodAnnotations returns the pod's annotations of given daemonSet
func GetDaemonSetPodAnnotations(item interface{}) map[string]string {
	return item.(appsv1.DaemonSet).Spec.Template.ObjectMeta.Annotations
}

// GetStatefulSetPodAnnotations returns the pod's annotations of given statefulSet
func GetStatefulSetPodAnnotations(item interface{}) map[string]string {
	return item.(appsv1.StatefulSet).Spec.Template.ObjectMeta.Annotations
}

// GetDeploymentVolumes returns the Volumes of given deployment
func GetDeploymentVolumes(item interface{}) []corev1.Volume {
	return item.(appsv1.Deployment).Spec.Template.Spec.Volumes
}

// GetDaemonSetVolumes returns the Volumes of given daemonSet
func GetDaemonSetVolumes(item interface{}) []corev1.Volume {
	return item.(appsv1.DaemonSet).Spec.Template.Spec.Volumes
}

// GetStatefulSetVolumes returns the Volumes of given statefulSet
func GetStatefulSetVolumes(item interface{}) []corev1.Volume {
	return item.(appsv1.StatefulSet).Spec.Template.Spec.Volumes
}

// GetDeploymentContainers returns the containers of given deployment
func GetDeploymentContainers(item interface{}) []corev1.Container {
	return item.(appsv1.Deployment).Spec.Template.Spec.Containers
}

// GetDaemonSetContainers returns the containers of given daemonSet
func GetDaemonSetContainers(item interface{}) []corev1.Container {
	return item.(appsv1.DaemonSet).Spec.Template.Spec.Containers
}

// GetStatefulSetContainers returns the containers of given statefulSet
func GetStatefulSetContainers(item interface{}) []corev1.Container {
	return item.(appsv1.StatefulSet).Spec.Template.Spec.Containers
}

// GetDeploymentInitContainers returns the containers of given deployment
func GetDeploymentInitContainers(item interface{}) []corev1.Container {
	return item.(appsv1.Deployment).Spec.Template.Spec.InitContainers
}

// GetDaemonSetInitContainers returns the containers of given daemonSet
func GetDaemonSetInitContainers(item interface{}) []corev1.Container {
	return item.(appsv1.DaemonSet).Spec.Template.Spec.InitContainers
}

// GetStatefulSetInitContainers returns the containers of given statefulSet
func GetStatefulSetInitContainers(item interface{}) []corev1.Container {
	return item.(appsv1.StatefulSet).Spec.Template.Spec.InitContainers
}

// UpdateDeployment performs rolling upgrade on deployment
func UpdateDeployment(client *kubernetes.Clientset, namespace string, resource interface{}) error {
	deployment := resource.(appsv1.Deployment)
	_, err := client.AppsV1().Deployments(namespace).Update(context.TODO(), &deployment, metav1.UpdateOptions{})
	return err
}

// UpdateDaemonSet performs rolling upgrade on daemonSet
func UpdateDaemonSet(client *kubernetes.Clientset, namespace string, resource interface{}) error {
	daemonSet := resource.(appsv1.DaemonSet)
	_, err := client.AppsV1().DaemonSets(namespace).Update(context.TODO(), &daemonSet, metav1.UpdateOptions{})
	return err
}

// UpdateStatefulSet performs rolling upgrade on statefulSet
func UpdateStatefulSet(client *kubernetes.Clientset, namespace string, resource interface{}) error {
	statefulSet := resource.(appsv1.StatefulSet)
	_, err := client.AppsV1().StatefulSets(namespace).Update(context.TODO(), &statefulSet, metav1.UpdateOptions{})
	return err
}

func updateEnvVar(containers []corev1.Container, envVar string, shaData string) Result {
	for i := range containers {
		envs := containers[i].Env
		for j := range envs {
			if envs[j].Name == envVar {
				if envs[j].Value != shaData {
					envs[j].Value = shaData
					return Updated
				}
				return NotUpdated
			}
		}
	}
	return NoEnvVarFound
}

// spec:
//   containers:
//     - name: test-container
//       image: k8s.gcr.io/busybox
//       env:
//         - name:
//           valueFrom:
//             configMapKeyRef:
//               name: special-config
//               key: SPECIAL_LEVEL
//       envFrom:
//         - configMapRef:
//             name: special-config
// 搜索resourceType/resourceName的第一个container
func getContainerWithEnvReference(containers []corev1.Container, resourceName string, resourceType string) *corev1.Container {
	for i := range containers {
		envs := containers[i].Env
		for j := range envs {
			envVarSource := envs[j].ValueFrom
			if envVarSource != nil {
				if resourceType == SecretEnvVarPostfix && envVarSource.SecretKeyRef != nil && envVarSource.SecretKeyRef.LocalObjectReference.Name == resourceName {
					return &containers[i]
				} else if resourceType == ConfigmapEnvVarPostfix && envVarSource.ConfigMapKeyRef != nil && envVarSource.ConfigMapKeyRef.LocalObjectReference.Name == resourceName {
					return &containers[i]
				}
			}
		}

		envsFrom := containers[i].EnvFrom
		for j := range envsFrom {
			if resourceType == SecretEnvVarPostfix && envsFrom[j].SecretRef != nil && envsFrom[j].SecretRef.LocalObjectReference.Name == resourceName {
				return &containers[i]
			} else if resourceType == ConfigmapEnvVarPostfix && envsFrom[j].ConfigMapRef != nil && envsFrom[j].ConfigMapRef.LocalObjectReference.Name == resourceName {
				return &containers[i]
			}
		}
	}
	return nil
}

// spec:
//   containers:
//     - name: test-container
//       image: k8s.gcr.io/busybox
//   volumes:
//     - name: config-volume
//       configMap:
//         name: special-config
//     - name: all-in-one
//       projected:
//         sources:
//           - secret:
//               name: user
func getVolumeMountName(volumes []corev1.Volume, mountType string, volumeName string) string {
	for i := range volumes {
		if mountType == ConfigmapEnvVarPostfix {
			if volumes[i].ConfigMap != nil && volumes[i].ConfigMap.Name == volumeName {
				return volumes[i].Name
			}

			if volumes[i].Projected != nil {
				for j := range volumes[i].Projected.Sources {
					if volumes[i].Projected.Sources[j].ConfigMap != nil && volumes[i].Projected.Sources[j].ConfigMap.Name == volumeName {
						return volumes[i].Name
					}
				}
			}
		} else if mountType == SecretEnvVarPostfix {
			if volumes[i].Secret != nil && volumes[i].Secret.SecretName == volumeName {
				return volumes[i].Name
			}

			if volumes[i].Projected != nil {
				for j := range volumes[i].Projected.Sources {
					if volumes[i].Projected.Sources[j].Secret != nil && volumes[i].Projected.Sources[j].Secret.Name == volumeName {
						return volumes[i].Name
					}
				}
			}
		}
	}

	return ""
}
func getContainerWithVolumeMount(containers []corev1.Container, volumeMountName string) *corev1.Container {
	for i := range containers {
		volumeMounts := containers[i].VolumeMounts
		for j := range volumeMounts {
			if volumeMounts[j].Name == volumeMountName {
				return &containers[i]
			}
		}
	}

	return nil
}

// 搜索containers里是否有container的volume和env使用了特定的configMap/secret
func getContainerToUpdate(upgradeFuncs RollingUpgradeFuncs, item interface{}, config Config, autoReload bool) *corev1.Container {
	volumes := upgradeFuncs.VolumesFunc(item)
	containers := upgradeFuncs.ContainersFunc(item)
	initContainers := upgradeFuncs.InitContainersFunc(item)

	var container *corev1.Container

	// Get the volumeMountName to find volumeMount in container
	volumeMountName := getVolumeMountName(volumes, config.Type, config.ResourceName)
	// Get the container with mounted configmap/secret
	if volumeMountName != "" {
		container = getContainerWithVolumeMount(containers, volumeMountName)
		if container == nil && len(initContainers) > 0 {
			container = getContainerWithVolumeMount(initContainers, volumeMountName)
			if container != nil {
				// if configmap/secret is being used in init container then return the first Pod container to save reloader env
				return &containers[0]
			}
		} else if container != nil {
			return container
		}
	}

	// Get the container with referenced secret or configmap as env var
	container = getContainerWithEnvReference(containers, config.ResourceName, config.Type)
	if container == nil && len(initContainers) > 0 {
		container = getContainerWithEnvReference(initContainers, config.ResourceName, config.Type)
		if container != nil {
			// if configmap/secret is being used in init container then return the first Pod container to save reloader env
			return &containers[0]
		}
	}

	// Get the first container if the annotation is related to specified configmap or secret i.e. configmap.reloader.stakater.com/reload
	if container == nil && !autoReload {
		return &containers[0]
	}

	return container
}

func updateContainers(upgradeFuncs RollingUpgradeFuncs, item interface{}, config Config, autoReload bool) Result {
	container := getContainerToUpdate(upgradeFuncs, item, config, autoReload)
	if container == nil {
		return NoContainerFound
	}

	var result Result
	envVar := EnvVarPrefix + util.ConvertToEnvVarName(config.ResourceName) + "_" + config.Type
	//update if env var exists
	result = updateEnvVar(upgradeFuncs.ContainersFunc(item), envVar, config.SHAValue)
	// if no existing env var exists lets create one
	if result == NoEnvVarFound {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  envVar,
			Value: config.SHAValue,
		})
		result = Updated
	}
	return result
}
func rollingUpgrade(clients *kubernetes.Clientset, config Config, upgradeFuncs RollingUpgradeFuncs, collectors metrics.Collectors) {
	items := upgradeFuncs.ItemsFunc(clients, config.Namespace)
	for _, item := range items {
		// find correct annotation and update the resource
		// 找resource annotation
		annotations := upgradeFuncs.AnnotationsFunc(item)
		annotationValue, found := annotations[config.Annotation]
		searchAnnotationValue, foundSearchAnn := annotations[options.AutoSearchAnnotation]
		reloaderEnabledValue, foundAuto := annotations[options.ReloaderAutoAnnotation]
		if !found && !foundAuto && !foundSearchAnn {
			// 找pod annotation
			annotations = upgradeFuncs.PodAnnotationsFunc(item)
			annotationValue = annotations[config.Annotation]
			searchAnnotationValue = annotations[options.AutoSearchAnnotation]
			reloaderEnabledValue = annotations[options.ReloaderAutoAnnotation]
		}

		result := NotUpdated
		reloaderEnabled, err := strconv.ParseBool(reloaderEnabledValue)
		if err == nil && reloaderEnabled {
			result = updateContainers(upgradeFuncs, item, config, true)
		}

		if result != Updated && annotationValue != "" {
			values := strings.Split(annotationValue, ",")
			for _, value := range values {
				value = strings.Trim(value, " ")
				if value == config.ResourceName {
					result = updateContainers(upgradeFuncs, item, config, false)
					if result == Updated {
						break
					}
				}
			}
		}

		if result != Updated && searchAnnotationValue == "true" {
			matchAnnotationValue := config.ResourceAnnotations[options.SearchMatchAnnotation]
			if matchAnnotationValue == "true" {
				result = updateContainers(upgradeFuncs, item, config, true)
			}
		}

		if result == Updated {
			// 重启deployment/daemonset/statefulset
			err = upgradeFuncs.UpdateFunc(clients, config.Namespace, item)
			resourceName := util.ToObjectMeta(item).Name
			if err != nil {
				logrus.Errorf("Update for '%s' of type '%s' named %s in namespace '%s' failed with error %v", resourceName, upgradeFuncs.ResourceType, config.ResourceName, config.Namespace, err)
				collectors.Counter.With(prometheus.Labels{"success": "false"}).Inc()
			} else {
				logrus.Infof("Changes detected in '%s' of type '%s' in namespace '%s'", config.ResourceName, config.Type, config.Namespace)
				logrus.Infof("Updated '%s' of type '%s' in namespace '%s'", resourceName, upgradeFuncs.ResourceType, config.Namespace)
				collectors.Counter.With(prometheus.Labels{"success": "true"}).Inc()
			}
		}
	}
}

func doRollingUpgrade(config Config, collectors metrics.Collectors) {
	clients, err := kube.GetKubernetesClient()
	if err != nil {
		logrus.Fatalf("Unable to create Kubernetes client error = %v", err)
	}

	rollingUpgrade(clients, config, GetDeploymentRollingUpgradeFuncs(), collectors)
	rollingUpgrade(clients, config, GetDaemonSetRollingUpgradeFuncs(), collectors)
	rollingUpgrade(clients, config, GetStatefulSetRollingUpgradeFuncs(), collectors)
}
