package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// debug: go run . --kubeconfig=`echo $HOME`/.kube/config
// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination
func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")

	var kubeconfig *string
	if home, _ := os.UserHomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	informerFactory := informers.NewSharedInformerFactory(k8sClient, time.Minute)

	podsQueue := workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(time.Second*5, time.Minute*5), "pods")
	endpointsQueue := workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(time.Second*5, time.Minute*5), "endpoints")

	informerFactory.Core().V1().Pods().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: nil,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod, ok := oldObj.(*corev1.Pod)
			if !ok {
				return
			}
			newPod, ok := newObj.(*corev1.Pod)
			if !ok {
				return
			}
			if oldPod.ResourceVersion == newPod.ResourceVersion {
				return
			}

			if newPod.Namespace != "default" {
				return
			}

			if newPod.DeletionTimestamp != nil {
				klog.Infof("pod %s/%s deleting", newPod.Namespace, newPod.Name)

				return
			}

			klog.Infof("pod %s/%s updated", newPod.Namespace, newPod.Name)
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return
			}

			klog.Infof("pod %s/%s deleted", pod.Namespace, pod.Name)

		},
	})

	informerFactory.Core().V1().Endpoints().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: nil,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldEndpoint, ok := oldObj.(*corev1.Endpoints)
			if !ok {
				return
			}
			newEndpoint, ok := newObj.(*corev1.Endpoints)
			if !ok {
				return
			}
			if newEndpoint.ResourceVersion == oldEndpoint.ResourceVersion {
				return
			}

			if newEndpoint.Namespace != "default" {
				return
			}

			klog.Infof("endpoints %s/%s updated", newEndpoint.Namespace, newEndpoint.Name)
		},
		DeleteFunc: func(obj interface{}) {
			endpoint, ok := obj.(*corev1.Endpoints)
			if !ok {
				return
			}

			klog.Infof("endpoints %s/%s deleted", endpoint.Namespace, endpoint.Name)

		},
	})

	stopCh := context.TODO().Done()

	informerFactory.Start(stopCh)

	defer podsQueue.ShutDown()
	defer endpointsQueue.ShutDown()

	informersSyncd := []cache.InformerSynced{
		informerFactory.Core().V1().Pods().Informer().HasSynced,
		informerFactory.Core().V1().Endpoints().Informer().HasSynced,
	}

	if !cache.WaitForCacheSync(stopCh, informersSyncd...) {
		klog.Errorf("Cannot sync pod, pv or pvc caches")
		return
	}

	go wait.Until(func() {
		key, quit := podsQueue.Get()
		if quit {
			return
		}
		defer podsQueue.Done(key)

		pod, ok := key.(*corev1.Pod)
		if !ok {
			return
		}

		klog.Infof("pod %s/%s deleted", pod.Namespace, pod.Name)
	}, time.Second, stopCh)

	go wait.Until(func() {
		key, quit := endpointsQueue.Get()
		if quit {
			return
		}
		defer endpointsQueue.Done(key)

		endpoints, ok := key.(*corev1.Endpoints)
		if !ok {
			return
		}

		klog.Infof("endpoints %s/%s updated", endpoints.Namespace, endpoints.Name)
	}, time.Second, stopCh)

	klog.Infof("cache synced...")

	<-stopCh
}
