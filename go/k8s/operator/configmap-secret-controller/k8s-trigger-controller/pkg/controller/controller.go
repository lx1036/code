package controller

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/metrics"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	DataHashAnnotation = "deployment.k8s.io/hash"

	ConfigMapsAnnotation = "k8s.io/configMaps"

	SecretsAnnotation = "k8s.io/secrets"
)

type Kind string

const (
	ConfigMap Kind = "configMap"
	Secret    Kind = "secret"
)

type item struct {
	object interface{}
	kind   Kind
	key    string
}

type Controller struct {
	client     *kubernetes.Clientset
	namespace  string
	collectors metrics.Collectors

	queue           workqueue.RateLimitingInterface
	informerFactory informers.SharedInformerFactory
}

var (
	accessor = meta.NewAccessor()
)

func NewController(informerFactory informers.SharedInformerFactory, client *kubernetes.Clientset, collectors metrics.Collectors, namespace string) (*Controller, error) {

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "configmap-secret")

	controller := &Controller{
		queue:           queue,
		informerFactory: informerFactory,
	}

	configMapInformer := controller.informerFactory.Core().V1().ConfigMaps()
	configMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.Enqueue(&item{
				object: obj,
				kind:   ConfigMap,
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			o, _ := accessor.ResourceVersion(oldObj.(runtime.Object))
			n, _ := accessor.ResourceVersion(newObj.(runtime.Object))
			// Only enqueue changes that have a different resource versions to avoid processing resyncs.
			if o != n {
				controller.Enqueue(&item{
					object: newObj,
					kind:   ConfigMap,
				})
			}
		},
	})

	secretInformer := controller.informerFactory.Core().V1().Secrets()
	secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.Enqueue(&item{
				object: obj,
				kind:   Secret,
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			o, _ := accessor.ResourceVersion(oldObj.(runtime.Object))
			n, _ := accessor.ResourceVersion(newObj.(runtime.Object))
			// Only enqueue changes that have a different resource versions to avoid processing resyncs.
			if o != n {
				controller.Enqueue(&item{
					object: newObj,
					kind:   Secret,
				})
			}
		},
	})

	deploymentInformer := controller.informerFactory.Apps().V1().Deployments()
	_ = deploymentInformer.Informer().AddIndexers(cache.Indexers{
		string(ConfigMap): controller.indexDeploymentsByConfigMap,
		string(Secret):    controller.indexDeploymentsBySecret,
	})

	return controller, nil
}

func (controller *Controller) indexDeploymentsByConfigMap(obj interface{}) ([]string, error) {
	var configs []string
	d, ok := obj.(*appsv1.Deployment)
	if !ok {
		return configs, fmt.Errorf("object is not deployment")
	}
	annotations := d.GetAnnotations()
	if len(annotations) == 0 {
		return configs, nil
	}
	if triggers, ok := annotations[ConfigMapsAnnotation]; ok {
		configMaps := sets.NewString(strings.Split(triggers, ",")...)
		// TODO: 这里搜索configmap不充分
		// 仅仅查看了volume https://kubernetes.io/docs/concepts/storage/volumes/#configmap
		// 还有container.env和container.envFrom作为环境变量
		for _, v := range d.Spec.Template.Spec.Volumes {
			if v.ConfigMap != nil {
				if configMaps.Has(v.ConfigMap.Name) {
					configs = append(configs, v.ConfigMap.Name)
				}
			}
		}
	}
	return configs, nil
}

func (controller *Controller) indexDeploymentsBySecret(obj interface{}) ([]string, error) {
	var configs []string
	d, ok := obj.(*appsv1.Deployment)
	if !ok {
		return configs, fmt.Errorf("object is not deployment")
	}
	annotations := d.GetAnnotations()
	if len(annotations) == 0 {
		return configs, nil
	}
	if triggers, ok := annotations[SecretsAnnotation]; ok {
		secrets := sets.NewString(strings.Split(triggers, ",")...)
		// TODO: 这里搜索configmap不充分
		// 仅仅查看了volume https://kubernetes.io/docs/concepts/storage/volumes/#configmap
		// 还有container.env和container.envFrom作为环境变量
		for _, v := range d.Spec.Template.Spec.Volumes {
			if v.Secret != nil {
				if secrets.Has(v.Secret.SecretName) {
					configs = append(configs, v.Secret.SecretName)
				}
			}
		}
	}
	return configs, nil
}

func (controller *Controller) Enqueue(item *item) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(item.object)
	if err != nil {
		return
	}
	item.key = key
	controller.queue.Add(item)
}

func (controller *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	go controller.informerFactory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh,
		controller.informerFactory.Core().V1().ConfigMaps().Informer().HasSynced,
		controller.informerFactory.Core().V1().Secrets().Informer().HasSynced,
		controller.informerFactory.Apps().V1().Deployments().Informer().HasSynced) {
		return fmt.Errorf("kubernetes informer unable to sync cache")
	}

	for i := 0; i < threadiness; i++ {
		// Wrap the process function with wait.Until so that if the controller crashes, it starts up again after a second.
		go wait.Until(func() {
			for controller.process() {
			}
		}, time.Second*1, stopCh)
	}

	return nil
}

// GenerateSHA generates SHA from string
func (controller *Controller) GenerateSHA(data string) string {
	hasher := sha1.New()
	_, err := io.WriteString(hasher, data)
	if err != nil {
		log.Errorf("unable to write data in hash writer %v", err)
	}
	sha := hasher.Sum(nil)
	return fmt.Sprintf("%x", sha)
}
func (controller *Controller) GetHashFromConfigmap(configMap *corev1.ConfigMap) string {
	var values []string
	for k, v := range configMap.Data {
		values = append(values, k+"="+v)
	}
	sort.Strings(values)
	return controller.GenerateSHA(strings.Join(values, ";"))
}
func (controller *Controller) GetHashFromSecret(secret *corev1.Secret) string {
	var values []string
	for k, v := range secret.Data {
		values = append(values, k+"="+string(v[:]))
	}
	sort.Strings(values)
	return controller.GenerateSHA(strings.Join(values, ";"))
}

func (controller *Controller) GetHash(obj interface{}) string {
	switch t := obj.(type) {
	case *corev1.ConfigMap:
		return controller.GetHashFromConfigmap(t)
	case *corev1.Secret:
		return controller.GetHashFromSecret(t)
	}

	utilruntime.HandleError(fmt.Errorf("unkown object type %T with: %v", obj, obj))
	return ""
}

func (controller *Controller) process() bool {
	keyObj, quit := controller.queue.Get()
	if quit {
		return false
	}

	err := func(obj interface{}) error {
		defer controller.queue.Done(obj)

		var entry *item
		var ok bool
		if entry, ok = obj.(*item); !ok {
			controller.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected *item in workqueue but got %#v", obj))
			return nil
		}

		var o interface{}
		switch entry.kind {
		case ConfigMap:
			_, exists, err := controller.informerFactory.Core().V1().ConfigMaps().Informer().GetStore().GetByKey(entry.key)
			if err != nil {
				return err
			}
			if !exists {
				log.Infof("object %+v was not found in the store", entry.key)
				return nil
			}

			//o, err = controller.informerFactory.Core().V1().ConfigMaps().Lister().ConfigMaps(entry.object.(*corev1.ConfigMap).Namespace).Get(entry.key)
			o, err = controller.client.CoreV1().ConfigMaps(entry.object.(*corev1.ConfigMap).Namespace).Get(context.TODO(), entry.key, metav1.GetOptions{})

			if err != nil {
				if errors.IsNotFound(err) {
					return nil
				}
				return err
			}

		case Secret:
			_, exists, err := controller.informerFactory.Core().V1().Secrets().Informer().GetStore().GetByKey(entry.key)
			if err != nil {
				return err
			}
			if !exists {
				log.Infof("object %+v was not found in the store", entry.key)
				return nil
			}

			//o, err = controller.informerFactory.Core().V1().Secrets().Lister().Secrets(entry.object.(*corev1.Secret).Namespace).Get(entry.key)
			o, err = controller.client.CoreV1().Secrets(entry.object.(*corev1.Secret).Namespace).Get(context.TODO(), entry.key, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					return nil
				}
				return err
			}
		}

		// First update the data hashes into ConfigMap/Secret annotations.
		objMeta, err := meta.Accessor(o)
		if err != nil {
			utilruntime.HandleError(err)
		}

		// Get all deployments that use the configMap or Secret
		filteringDeployments, err := controller.informerFactory.Apps().V1().Deployments().Informer().GetIndexer().ByIndex(string(entry.kind), objMeta.GetName())
		if err != nil {
			return err
		}
		if len(filteringDeployments) == 0 {
			log.Debugf("has no deployment use %s: %s/%s", entry.kind, objMeta.GetNamespace(), objMeta.GetName())
			return nil
		}

		// get hash of configMap or secret, and update new hash data in annotation key
		hashVersion := controller.GetHash(o)
		if len(hashVersion) == 0 {
			return nil
		}

		oldHashVersion := objMeta.GetAnnotations()[DataHashAnnotation]
		if hashVersion != oldHashVersion {
			switch entry.kind {
			case ConfigMap:
				c := o.(*corev1.ConfigMap).DeepCopy()
				if c.Annotations == nil {
					c.Annotations = map[string]string{}
				}
				c.Annotations[DataHashAnnotation] = hashVersion
				log.Infof("updating %s %s/%s with new data hash: %s", c.Kind, c.Namespace, c.Name, hashVersion)
				if _, err := controller.client.CoreV1().ConfigMaps(c.Namespace).Update(context.TODO(), c, metav1.UpdateOptions{}); err != nil {
					if errors.IsNotFound(err) {
						return nil
					}
					return err
				}
			case Secret:
				c := o.(*corev1.Secret).DeepCopy()
				if c.Annotations == nil {
					c.Annotations = map[string]string{}
				}
				c.Annotations[DataHashAnnotation] = hashVersion
				log.Infof("updating %s %s/%s with new data hash: %s", c.Kind, c.Namespace, c.Name, hashVersion)
				if _, err := controller.client.CoreV1().Secrets(c.Namespace).Update(context.TODO(), c, metav1.UpdateOptions{}); err != nil {
					if errors.IsNotFound(err) {
						return nil
					}
					return err
				}
			}
		} else {
			log.Infof("no change detected in hash for %s %s", entry.kind, entry.key)
		}

		// check if rolling update deployments
		for _, obj := range filteringDeployments {
			deploymentMeta, err := meta.Accessor(obj)
			if err != nil {
				utilruntime.HandleError(err)
				continue
			}

			deployment, err := controller.client.AppsV1().Deployments(deploymentMeta.GetNamespace()).Get(context.TODO(), deploymentMeta.GetName(), metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					log.Errorf("not found deployment %s/%s", deployment.Namespace, deployment.Name)
					continue
				}

				log.Errorf("get deployment %s/%s with err: %v", deployment.Namespace, deployment.Name, err)
				continue
			}

			annotations := deployment.Spec.Template.Annotations
			if annotations == nil {
				annotations = map[string]string{}
			}
			annotationKey := fmt.Sprintf("k8s.io/%s-%s", entry.kind, objMeta.GetName())
			if annotations[annotationKey] != hashVersion {
				// rolling update deployment
				c := deployment.DeepCopy()
				if c.Annotations == nil {
					c.Annotations = map[string]string{}
				}
				c.Spec.Template.Annotations[annotationKey] = hashVersion
				if _, err := controller.client.AppsV1().Deployments(c.Namespace).Update(context.TODO(), c, metav1.UpdateOptions{}); err != nil {
					log.Errorf("failed to update deployment %s/%s: %v", c.Namespace, c.Name, err)
					continue
				}
			}
		}

		return nil
	}(keyObj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}
