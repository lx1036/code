package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/pkg/metrics"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const Name = "filebeat"

const (
	LogEnableAnnotation = "k8s.io/log-enable"
	LogConfigAnnotation = "k8s.io/log-config"
)

type LogController struct {
	InformerResources []schema.GroupVersionResource
	ApiServerClient   kubernetes.Interface
	PodStore          cache.Store
	PodInformer       cache.Controller
	NodeStore         cache.Store
	NodeInformer      cache.Controller

	TaskQueue *TaskQueue

	stopCh chan struct{}

	NodeName         string
	ResyncPeriod     time.Duration
	TaskHandlePeriod time.Duration
}

var (
	accessor = meta.NewAccessor()
)

const defaultNode = "localhost"

func InClusterNamespace() (string, error) {
	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func DiscoverKubernetesNode(host string, client kubernetes.Interface) string {
	if len(host) != 0 {
		log.Infof("Using node %s provided by env", host)
		return host
	}

	// node discover by pod
	ns, err := InClusterNamespace()
	if err != nil {
		log.Errorf("Can't get namespace in cluster with error: %v", err)
		return defaultNode
	}
	podName, err := os.Hostname()
	if err != nil {
		log.Errorf("Can't get hostname as pod name in cluster with error: %v", err)
		return defaultNode
	}

	pod, err := client.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Can't query pod in cluster with error: %v", err)
		return defaultNode
	}

	log.Infof("Using node %s discovered by pod in cluster", pod.Spec.NodeName)

	return pod.Spec.NodeName
}

type Controller struct {
	queue           workqueue.RateLimitingInterface
	informerFactory informers.SharedInformerFactory
	client          *kubernetes.Clientset

	podInformer  cache.SharedIndexInformer
	nodeInformer cache.SharedIndexInformer

	cache *Cache
}

// Resource data
type Resource = runtime.Object
type WatchOptions struct {
	Namespace string
	Node      string

	// SyncTimeout is timeout for listing historical resources
	SyncTimeout time.Duration
}

func nodeSelector(options *metav1.ListOptions, opt WatchOptions) {
	if len(opt.Node) != 0 {
		options.FieldSelector = fmt.Sprintf("spec.nodeName=%s", opt.Node)
	}
}
func nameSelector(options *metav1.ListOptions, opt WatchOptions) {
	if len(opt.Node) != 0 {
		options.FieldSelector = fmt.Sprintf("metadata.name=%s", opt.Node)
	}
}
func NewInformer(client kubernetes.Interface, resource Resource, opts WatchOptions, indexers cache.Indexers) cache.SharedIndexInformer {
	ctx := context.TODO()
	var listWatch *cache.ListWatch
	switch resource.(type) {
	case *coreV1.Pod:
		pod := client.CoreV1().Pods(opts.Namespace)
		listWatch = &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				nodeSelector(&options, opts)
				return pod.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				nodeSelector(&options, opts)
				return pod.Watch(ctx, options)
			},
		}
	case *coreV1.Node:
		node := client.CoreV1().Nodes()
		listWatch = &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				nameSelector(&options, opts)
				return node.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				nameSelector(&options, opts)
				return node.Watch(ctx, options)
			},
		}
	}

	return cache.NewSharedIndexInformer(listWatch, resource, opts.SyncTimeout, indexers)
}

func NewController(informerFactory informers.SharedInformerFactory, client *kubernetes.Clientset, collectors metrics.Collectors) (*Controller, error) {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), Name)
	controller := &Controller{
		queue:           queue,
		informerFactory: informerFactory,
		client:          client,
		cache:           newCache(),
	}

	podInformer := NewInformer(client, &coreV1.Pod{}, WatchOptions{
		Namespace:   viper.GetString("namespace"),
		Node:        DiscoverKubernetesNode(viper.GetString("node"), client),
		SyncTimeout: viper.GetDuration("sync-period"),
	}, cache.Indexers{})
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.AddPod,
		UpdateFunc: controller.UpdatePod,
		DeleteFunc: controller.DeletePod,
	})
	controller.podInformer = podInformer

	nodeInformer := NewInformer(client, &coreV1.Node{}, WatchOptions{
		Node:        DiscoverKubernetesNode(viper.GetString("node"), client),
		SyncTimeout: viper.GetDuration("sync-period"),
	}, cache.Indexers{})

	controller.nodeInformer = nodeInformer

	return controller, nil
}

type Action string

const (
	Add    Action = "add"
	Update Action = "update"
	Delete Action = "delete"
)

type item struct {
	object interface{}
	key    string
	action Action
}

func (controller *Controller) UpdatePod(oldObj, newObj interface{}) {
	o, _ := accessor.ResourceVersion(oldObj.(runtime.Object))
	n, _ := accessor.ResourceVersion(newObj.(runtime.Object))
	// 只有resource version不同才是新对象
	if o != n {
		controller.Enqueue(&item{
			object: newObj,
			action: Update,
		})
	}
}

func (controller *Controller) AddPod(obj interface{}) {
	controller.Enqueue(&item{
		object: obj,
		action: Add,
	})
}

func (controller *Controller) DeletePod(obj interface{}) {
	controller.Enqueue(&item{
		object: obj,
		action: Delete,
	})
}

func (controller *Controller) Enqueue(item *item) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(item.object)
	if err != nil {
		log.Errorf("fail to get key for %v", item.object)
		return
	}

	// filter pod without log configuration
	pod, ok := item.object.(*coreV1.Pod)
	if !ok {
		log.Errorf("expected *coreV1.Pod but got %T", item.object)
		return
	}

	if value, ok := pod.Annotations[LogEnableAnnotation]; !ok || value != "true" {
		return
	}

	if _, ok := pod.Annotations[LogConfigAnnotation]; !ok {
		return
	}

	// only queue ready pod, 同一个pod在启动过程中，会出现多种状态，直至最后status.conditions都是ready状态
	// 但是会触发多次的update event
	for _, condition := range pod.Status.Conditions {
		if condition.Type == coreV1.PodReady && condition.Status != coreV1.ConditionTrue {
			return
		}
	}

	item.key = key
	controller.queue.Add(item)
}

func (controller *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	go controller.podInformer.Run(stopCh)
	go controller.nodeInformer.Run(stopCh)

	if !cache.WaitForNamedCacheSync(Name, stopCh,
		controller.podInformer.HasSynced,
		controller.nodeInformer.HasSynced) {
		return fmt.Errorf("kubernetes informer is unable to sync cache")
	}

	for i := 0; i < threadiness; i++ {
		// Wrap the process function with wait.Until so that if the controller crashes, it starts up again after a second.
		go wait.Until(func() {
			for controller.process() {
			}
		}, time.Second*1, stopCh)
	}

	go wait.Until(controller.syncFilebeatYamlByIncremental, time.Second*30, stopCh)

	go wait.Until(controller.syncFilebeatYamlByTotal, viper.GetDuration("sync-period"), stopCh)

	return nil
}

type MultilineConfig struct {
	MulPattern string `json:"multiline_pattern"`
	MulNegate  string `json:"multiline_negate"`
	MulMatch   string `json:"multiline_match"`
}
type LogCollectorType string

const (
	Daemonset LogCollectorType = "daemonset"
	Sidecar   LogCollectorType = "sidecar"
)

type LogType string

const (
	Stdout LogType = "stdout"
	File   LogType = "file"
)

type ContainerLogConfig struct {
	LogCollectorType LogCollectorType `json:"log_collector_type"` //0: sidecar, 1: daemonset
	LogType          LogType          `json:"log_type,omitempty"` //0: stdout， 1: filelog
	Topic            string           `json:"topic"`
	Hosts            string           `json:"hosts"`
	Containers       []string         `json:"containers"`

	Paths            []string        `json:"paths,omitempty"` //only daemonset mode, and stdout paths will be nil
	MultilineEnable  bool            `json:"multiline_enable"`
	MultilinePattern MultilineConfig `json:"multiline_pattern,omitempty"`
}

//type ContainerLogConfigs map[string]ContainerLogConfig // key写container name
type Conf struct {
	ContainerLogConfigs ContainerLogConfig `json:"containerLogConfigs"`
}

func decodeConfig(config string) (*Conf, error) {
	conf := &Conf{}
	if c := strings.TrimSpace(config); len(c) != 0 {
		err := json.Unmarshal([]byte(c), conf)
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}

type FilebeatInput struct {
	Hosts            string
	Paths            []string
	HostsTopic       string
	Topic            string
	MultilineEnable  bool
	MultilinePattern MultilineConfig
	CustomField      string
}

func getContainerLogPath() string {
	logPath := viper.GetString("container-log-path")
	if logPath[len(logPath)-1:] != string(os.PathSeparator) {
		logPath += string(os.PathSeparator)
	}

	return logPath
}
func inSlice(slices []string, element string) bool {
	for _, slice := range slices {
		if slice == element {
			return true
		}
	}

	return false
}

type FilebeatInputsData struct {
	FilebeatInputs []FilebeatInput `json:"FilebeatInputs"`
}

func (controller *Controller) syncFilebeatYamlByIncremental() {
	var inputs []FilebeatInput

	if !controller.cache.changed { // cache没有变化
		return
	}

	log.Info("calculating filebeat input data...")

	for key, pod := range controller.cache.items {
		if config := strings.TrimSpace(pod.Annotations[LogConfigAnnotation]); len(config) != 0 {
			conf, err := decodeConfig(config)
			if err != nil {
				log.Errorf("fail to decode json config with pod %s", key)
				continue
			}
			switch conf.ContainerLogConfigs.LogCollectorType {
			case Daemonset:
				switch conf.ContainerLogConfigs.LogType {
				case Stdout:
					input := FilebeatInput{
						Hosts:           conf.ContainerLogConfigs.Hosts,
						HostsTopic:      fmt.Sprintf("%s-%s", conf.ContainerLogConfigs.Hosts, conf.ContainerLogConfigs.Topic),
						Topic:           conf.ContainerLogConfigs.Topic,
						MultilineEnable: conf.ContainerLogConfigs.MultilineEnable,
						CustomField:     fmt.Sprintf("IDC=%s,PodName=%s", viper.GetString("IDC"), pod.Name),
					}
					if input.MultilineEnable {
						input.MultilinePattern = conf.ContainerLogConfigs.MultilinePattern
					}

					containerLogPath := getContainerLogPath()
					for _, containerStatus := range pod.Status.ContainerStatuses {
						if !inSlice(conf.ContainerLogConfigs.Containers, containerStatus.Name) {
							continue
						}

						containerId := strings.TrimPrefix(containerStatus.ContainerID, "docker://")
						input.Paths = append(input.Paths, fmt.Sprintf("%s%s/%s-json.log", containerLogPath, containerId, containerId))

					}

					inputs = append(inputs, input)
				case File:

				default:
					log.Errorf("unsupported log type for pod %s", key)
				}
			case Sidecar:

			default:
				log.Errorf("unsupported log collector type for pod %s", key)
			}
		}
	}

	log.Info("syncing filebeat inputs.yml by data...")

	filebeatInputsData := FilebeatInputsData{
		FilebeatInputs: inputs,
	}

	tpl, err := template.ParseFiles("config/inputs.yml.template")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	buf := bytes.NewBufferString("")
	if err = tpl.Execute(buf, filebeatInputsData); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if err = ioutil.WriteFile("inputs.yml", buf.Bytes(), os.ModePerm); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Info("synced filebeat inputs.yml by data")

	controller.cache.changed = false
}

func (controller *Controller) syncFilebeatYamlByTotal() {

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

		obj, exists, err := controller.podInformer.GetStore().GetByKey(entry.key)
		if err != nil {
			return err
		}
		if !exists {
			log.Infof("object %+v was not found in the store", entry.key)
			return nil
		}

		var pod *coreV1.Pod
		if pod, ok = obj.(*coreV1.Pod); !ok {
			controller.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected *coreV1.Pod but got %#v", obj))
			return nil
		}

		switch entry.action {
		case Add:
			log.Infof("adding cache with pod %s/%s", pod.Namespace, pod.Name)
			controller.cache.Set(entry.key, pod)
		case Update:
			// INFO: 多次 {"level":"info","msg":"updating cache with pod default/daemonset-stdout-demo-lx1036-7d8c79f474-zwfzm","time":"2020-12-06T18:10:17+08:00"}
			// 为何这么多次？
			// 当pod没有ready时，会多次触发update event，queue里就会多次有update pod加进来
			// 所以需要过滤，只有ready状态的pod进入queue
			log.Infof("updating cache with pod %s/%s", pod.Namespace, pod.Name)
			controller.cache.Set(entry.key, pod)
		case Delete:
			// INFO: 当kubectl delete -f deployment.yaml删除deployment时候，居然触发的是update事件
			// INFO: 现在问题是：删除pod后，inputs.yml文件并没有立即删除pod对应的配置(或许应该这样设计，交给捞底sync去做？)
			log.Infof("deleting cache with pod %s/%s", pod.Namespace, pod.Name)
			controller.cache.Delete(entry.key)
		}

		return nil
	}(keyObj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}
