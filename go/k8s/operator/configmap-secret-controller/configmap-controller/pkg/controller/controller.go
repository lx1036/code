package controller

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/golang/glog"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/configmap-controller/pkg/kube"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/configmap-controller/pkg/metrics"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/handler"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/util"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"sort"
	"strings"
	"time"
)

const (
	UpdateOnChangeAnnotation = "configmap.lx1036.io/update-on-change"
	Separator                = ","
)

// Controller for checking events
type Controller struct {
	//indexer    cache.Indexer
	//queue      workqueue.RateLimitingInterface
	//informer   cache.Controller

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
		log.Errorf("Couldn't create kubernetes watcher for %T", &ConfigMap{})
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

// updateContainers returns a boolean value indicating if any containers have been updated
func updateContainers(containers []api.Container, annotationValue, configMapVersion string) bool {

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
		log.Errorf("Unable to write data in hash writer %v", err)
	}
	sha := hasher.Sum(nil)
	return fmt.Sprintf("%x", sha)
}

func (controller *Controller) rollingUpdateDeployment(configMap *ConfigMap, action string) {
	deploymentList, err := controller.client.AppsV1().Deployments(configMap.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {

		return
	}

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
						log.Errorf("update deployment error: %s", err)
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
	/*defer runtime.HandleCrash()
	// Let the workers stop when we are done
	defer c.queue.ShutDown()

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForNamedCacheSync(c.resource, stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}*/

	if err := controller.watcher.Start(); err != nil {
		log.Debugf("add_kubernetes_metadata", "Couldn't start watcher: %v", err)
		return
	}

	<-stopCh
	log.Infof("Stopping Controller")
}
