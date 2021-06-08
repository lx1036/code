package main

import (
	"context"
	"flag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"time"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "kubeconfig path")
)

const (
	evictionWatchRetryWait    = 10 * time.Second
	evictionWatchJitterFactor = 0.5
)

// INFO: 这个逻辑可以复用：watch event 事件，如果 watch 过程中有失败，则等待 jitter 时间重新去 watch

// go run . --kubeconfig=`echo $HOME/.kube/config`
func main() {
	flag.Parse()

	if len(*kubeconfig) == 0 {
		os.Exit(1)
	}
	config, err := NewRestConfig(*kubeconfig)
	if err != nil {
		panic(err)
	}

	kubeClient := kubernetes.NewForConfigOrDie(config)
	stopCh := genericapiserver.SetupSignalHandler()

	go func() {
		watchEvictionEventsOnce := func() {
			watcher, err := kubeClient.CoreV1().Events(metav1.NamespaceAll).Watch(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Error(err)
				return
			}
			// 读取 event 数据
			for {
				evictedEvent, ok := <-watcher.ResultChan() // 这里 channel 阻塞，会一直 watch 新的 event 到来
				if !ok {
					klog.V(3).Infof("Eviction event chan closed")
					return
				}
				evictedEvent2, ok := evictedEvent.Object.(*corev1.Event)
				if !ok {
					return
				}

				klog.Info(evictedEvent2.String())
			}
		}

		// INFO: 这里这个逻辑可以复用
		for {
			watchEvictionEventsOnce() // 如果触发了 return 逻辑，则重新去 watch，确保 watch 是正常工作的
			// INFO: 如果 watch 失败，等待 jitter 时间后重新 watch，确保 watch 是正常工作的
			waitTime := wait.Jitter(evictionWatchRetryWait, evictionWatchJitterFactor) // 10 + 10 * 0.5 * rand[0,1)
			time.Sleep(waitTime)
		}
	}()

	<-stopCh
}

func NewRestConfig(kubeconfig string) (*rest.Config, error) {
	var config *rest.Config
	if _, err := os.Stat(kubeconfig); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	} else { //Use Incluster Configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	// Use protobufs for communication with apiserver
	//config.ContentType = "application/vnd.kubernetes.protobuf"
	return config, nil
}
