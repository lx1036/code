package controller

import (
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/metrics"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	DataHashAnnotation = "k8s.io/hash"

	ConfigMapsAnnotation = "k8s.io/configMaps"

	SecretsAnnotation = "k8s.io/secrets"

	ReloadAnnotation = "k8s.io/reload"
)

type Kind string

const (
	ConfigMap Kind = "configMap"
	Secret    Kind = "secret"
)

const Name = "configmap-secret"

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
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), Name)

	controller := &Controller{
		queue:           queue,
		informerFactory: informerFactory,
		client:          client,
		collectors:      collectors,
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
			// 只有resource version不同才是新对象
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
		// volume使用configmap
		for _, v := range d.Spec.Template.Spec.Volumes {
			if v.ConfigMap != nil {
				if configMaps.Has(v.ConfigMap.Name) {
					configs = append(configs, v.ConfigMap.Name)
				}
			}
		}

		// env使用configmap
		for _, container := range d.Spec.Template.Spec.Containers {
			for _, envVar := range container.Env {
				if envVar.ValueFrom != nil && envVar.ValueFrom.ConfigMapKeyRef != nil && configMaps.Has(envVar.ValueFrom.ConfigMapKeyRef.Name) {
					configs = append(configs, envVar.ValueFrom.ConfigMapKeyRef.Name)
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
		// volume使用secret
		for _, v := range d.Spec.Template.Spec.Volumes {
			if v.Secret != nil {
				if secrets.Has(v.Secret.SecretName) {
					configs = append(configs, v.Secret.SecretName)
				}
			}
		}

		// env使用secret
		for _, container := range d.Spec.Template.Spec.Containers {
			for _, envVar := range container.Env {
				if envVar.ValueFrom != nil && envVar.ValueFrom.SecretKeyRef != nil && secrets.Has(envVar.ValueFrom.SecretKeyRef.Name) {
					configs = append(configs, envVar.ValueFrom.SecretKeyRef.Name)
				}
			}
		}
	}
	return configs, nil
}

func (controller *Controller) Enqueue(item *item) {
	annotations, err := accessor.Annotations(item.object.(runtime.Object))
	if err != nil {
		return
	}
	if value, ok := annotations[ReloadAnnotation]; !ok || value != "true" {
		return
	}

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(item.object)
	if err != nil {
		return
	}
	item.key = key

	controller.queue.Add(item)
}

func (controller *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	go controller.informerFactory.Start(stopCh)

	if !cache.WaitForNamedCacheSync(Name, stopCh,
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

			// 这只是从本地缓存local storage里取数据
			o, err = controller.informerFactory.Core().V1().ConfigMaps().Lister().ConfigMaps(entry.object.(*corev1.ConfigMap).Namespace).Get(entry.object.(*corev1.ConfigMap).Name)
			// http get请求api-server数据
			//o, err = controller.client.CoreV1().ConfigMaps(entry.object.(*corev1.ConfigMap).Namespace).Get(context.TODO(), entry.key, metav1.GetOptions{})
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

			o, err = controller.informerFactory.Core().V1().Secrets().Lister().Secrets(entry.object.(*corev1.Secret).Namespace).Get(entry.object.(*corev1.Secret).Name)
			//o, err = controller.client.CoreV1().Secrets(entry.object.(*corev1.Secret).Namespace).Get(context.TODO(), entry.key, metav1.GetOptions{})
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

		// 从本地缓存里查找使用该configMap/secret的所有deployments
		// 参考L117-L120
		filteringDeployments, err := controller.informerFactory.Apps().V1().Deployments().Informer().GetIndexer().ByIndex(string(entry.kind), objMeta.GetName())
		if err != nil {
			return err
		}
		if len(filteringDeployments) == 0 {
			log.Debugf("has no deployment use %s: %s/%s", entry.kind, objMeta.GetNamespace(), objMeta.GetName())
			return nil
		}

		// 求算configMap/secret的data的hash值
		hashVersion := controller.GetHash(o)
		if len(hashVersion) == 0 {
			return nil
		}

		oldHashVersion := objMeta.GetAnnotations()[DataHashAnnotation]
		if hashVersion != oldHashVersion {
			// configMap/secret的data的hash不一样，说明数据已经修改了
			switch entry.kind {
			case ConfigMap:
				c := o.(*corev1.ConfigMap).DeepCopy()
				if c.Annotations == nil {
					c.Annotations = map[string]string{}
				}
				c.Annotations[DataHashAnnotation] = hashVersion
				log.Infof("updating %s %s/%s with new data hash: %s", entry.kind, c.Namespace, c.Name, hashVersion)
				// http请求去更新configmap
				// 这里是不是可以使用patch()更好些？
				if _, err := controller.client.CoreV1().ConfigMaps(c.Namespace).Update(context.TODO(), c, metav1.UpdateOptions{}); err != nil {
					controller.collectors.ConfigMapCounter.With(prometheus.Labels{"configMap": "fail"}).Inc()
					if errors.IsNotFound(err) {
						return nil
					}
					return err
				}
				controller.collectors.ConfigMapCounter.With(prometheus.Labels{"configMap": "success"}).Inc()
			case Secret:
				c := o.(*corev1.Secret).DeepCopy()
				if c.Annotations == nil {
					c.Annotations = map[string]string{}
				}
				c.Annotations[DataHashAnnotation] = hashVersion
				log.Infof("updating %s %s/%s with new data hash: %s", entry.kind, c.Namespace, c.Name, hashVersion)
				if _, err := controller.client.CoreV1().Secrets(c.Namespace).Update(context.TODO(), c, metav1.UpdateOptions{}); err != nil {
					controller.collectors.SecretCounter.With(prometheus.Labels{"secret": "fail"}).Inc()
					if errors.IsNotFound(err) {
						return nil
					}
					return err
				}
				controller.collectors.SecretCounter.With(prometheus.Labels{"secret": "success"}).Inc()
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
				log.Infof("updating deployment %s/%s pod-template annotations with hash %s", deployment.Namespace, deployment.Name, hashVersion)
				c := deployment.DeepCopy()
				if c.Spec.Template.Annotations == nil {
					c.Spec.Template.Annotations = map[string]string{}
				}
				c.Spec.Template.Annotations[annotationKey] = hashVersion
				// http更新deployment
				if _, err := controller.client.AppsV1().Deployments(c.Namespace).Update(context.TODO(), c, metav1.UpdateOptions{}); err != nil {
					controller.collectors.DeploymentCounter.With(prometheus.Labels{"deployment": "fail"}).Inc()
					log.Errorf("failed to update deployment %s/%s: %v", c.Namespace, c.Name, err)
					continue
				}
				controller.collectors.DeploymentCounter.With(prometheus.Labels{"deployment": "success"}).Inc()
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
