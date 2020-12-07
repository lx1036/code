package controller

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"k8s-lx1036/k8s/operator/configmap-secret-controller/configmap-controller/pkg/metrics"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	UpdateOnChangeAnnotation = "configmap.lx1036.io/update-on-change"
	Separator                = ","
)

// Controller for checking events
type Controller struct {
	watcher    Watcher
	client     *kubernetes.Clientset
	namespace  string
	resource   string
	collectors metrics.Collectors
}

// NewController for initializing a Controller
func NewController(client *kubernetes.Clientset, resource string, namespace string, collectors metrics.Collectors) (*Controller, error) {
	controller := &Controller{
		client:     client,
		namespace:  namespace,
		resource:   resource,
		collectors: collectors,
	}

	watcher, err := NewWatcher(client, &ConfigMap{}, WatchOptions{
		SyncTimeout: viper.GetDuration("sync-period"),
		Namespace:   namespace,
	}, nil)
	if err != nil {
		log.Errorf("unable to create kubernetes watcher for %T", &ConfigMap{})
		return nil, err
	}

	watcher.AddEventHandler(ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			configMap := obj.(*ConfigMap)
			log.Debugf("Adding kubernetes configmap: %s/%s", configMap.GetNamespace(), configMap.GetName())
			controller.onAdd(configMap)
		},
		UpdateFunc: func(obj interface{}) {
			configMap := obj.(*ConfigMap)
			log.Debugf("Updating kubernetes configmap: %s/%s", configMap.GetNamespace(), configMap.GetName())
			controller.onUpdate(configMap)
		},
		DeleteFunc: nil,
	})

	controller.watcher = watcher

	return controller, nil
}

// convertToEnvVarName converts the given text into a usable env var
// removing any special chars with '_'
func convertToEnvVarName(text string) string {
	var buffer bytes.Buffer
	upper := strings.ToUpper(text)
	lastCharValid := false
	for i := 0; i < len(upper); i++ {
		ch := upper[i]
		if ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' {
			buffer.WriteByte(ch)
			lastCharValid = true
			continue
		}
		if lastCharValid {
			buffer.WriteByte('_')
		}
		lastCharValid = false
	}
	return buffer.String()
}

// updateContainers returns a boolean value indicating if any containers have been updated
func updateContainers(containers []corev1.Container, annotationValue, configMapVersion string) bool {
	// we can have multiple configmaps to update
	updated := false
	configmaps := strings.Split(annotationValue, Separator)
	for _, configmap := range configmaps {
		customEnvName := "PREFIX_" + convertToEnvVarName(configmap)
		for _, container := range containers {
			envs := container.Env
			matched := false
			for _, env := range envs {
				if env.Name == customEnvName {
					matched = true
					if env.Value != configMapVersion {
						log.Infof("Updating %s to %s", customEnvName, configMapVersion)
						env.Value = configMapVersion
						updated = true
					}
				}
			}

			// if no existing env var exists lets create one
			if !matched {
				e := corev1.EnvVar{
					Name:  customEnvName,
					Value: configMapVersion,
				}
				container.Env = append(container.Env, e)
				updated = true
			}
		}
	}

	return updated
}

func GetHashFromConfigmap(configMap *ConfigMap) string {
	var values []string
	for k, v := range configMap.Data {
		values = append(values, k+"="+v)
	}
	sort.Strings(values)
	return GenerateSHA(strings.Join(values, ";"))
}

// GenerateSHA generates SHA from string
func GenerateSHA(data string) string {
	hasher := sha1.New()
	_, err := io.WriteString(hasher, data)
	if err != nil {
		log.Errorf("unable to write data in hash writer %v", err)
	}
	sha := hasher.Sum(nil)
	return fmt.Sprintf("%x", sha)
}

func (controller *Controller) rollingUpdateDeployment(configMap *ConfigMap, action string) {
	// TODO: 这里list所有的deployments，并不好
	deploymentList, err := controller.client.AppsV1().Deployments(configMap.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("unable to list deployment in namespace %s with %v", configMap.Namespace, err)
		return
	}

	// hash下configmap.data，与新的configMap的hash进行比较，来判断是否需要重启deployment
	configMapVersion := GetHashFromConfigmap(configMap)

	for _, deployment := range deploymentList.Items {
		annotationValue := deployment.ObjectMeta.Annotations[UpdateOnChangeAnnotation]
		if len(annotationValue) != 0 {
			values := strings.Split(annotationValue, Separator)
			matches := false
			for _, value := range values {
				if value == configMap.Name {
					matches = true
					break
				}
			}
			if matches {
				if updateContainers(deployment.Spec.Template.Spec.Containers, annotationValue, configMapVersion) {
					// update the deployment
					_, err := controller.client.AppsV1().Deployments(configMap.Namespace).Update(context.TODO(), &deployment, metav1.UpdateOptions{})
					if err != nil {
						log.Errorf("update deployment error: %v", err)
						controller.watcher.EnqueueAfter(configMap, action, time.Second*3)
						return
					}
					log.Infof("Updated Deployment %s", deployment.Name)
				}
			}
		}
	}
}

func (controller *Controller) onAdd(configMap *ConfigMap) {
	controller.rollingUpdateDeployment(configMap, Add)
}

func (controller *Controller) onUpdate(configMap *ConfigMap) {
	controller.rollingUpdateDeployment(configMap, Update)
}

//Run function for controller which handles the queue
func (controller *Controller) Run(threadiness int, stopCh chan struct{}) {
	if err := controller.watcher.Start(threadiness); err != nil {
		log.Errorf("unable to start watcher: %v", err)
		return
	}

	<-stopCh
	log.Infof("Stopping Controller")
}
