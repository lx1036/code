package controller

import (
	"context"
	"fmt"
	gobgpapi "github.com/osrg/gobgp/v3/api"
	v1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/apis/bgplb.k9s.io/v1"
	"k8s.io/client-go/tools/cache"
	"net"
	"time"

	"k8s-lx1036/k8s/network/loadbalancer/bgplb/cmd/app/options"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/clientset/versioned"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/informers/externalversions"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/utils"

	gobgp "github.com/osrg/gobgp/v3/pkg/server"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type BgpLBController struct {
	bgpServer *gobgp.BgpServer

	bgpConfInformer cache.SharedIndexInformer
	bgpPeerInformer cache.SharedIndexInformer
	bgpEipInformer  cache.SharedIndexInformer
}

func NewController(option *options.Options) (*BgpLBController, error) {
	maxSize := 4 << 20 //4MB
	grpcOpts := []grpc.ServerOption{grpc.MaxRecvMsgSize(maxSize), grpc.MaxSendMsgSize(maxSize)}
	bgpServer := gobgp.NewBgpServer(gobgp.GrpcListenAddress(option.GrpcHosts), gobgp.GrpcOption(grpcOpts))

	restConfig, err := utils.NewRestConfig(option.Kubeconfig)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}

	var bgplbFactoryOpts []externalversions.SharedInformerOption
	if option.Namespace != corev1.NamespaceAll {
		bgplbFactoryOpts = append(bgplbFactoryOpts, externalversions.WithNamespace(option.Namespace))
	}

	bgplbClient := versioned.NewForConfigOrDie(restConfig)
	bgplbInformerFactory := externalversions.NewSharedInformerFactoryWithOptions(bgplbClient, time.Second*30, bgplbFactoryOpts...)
	bgpConfInformer := bgplbInformerFactory.Bgplb().V1().BgpConves().Informer()
	bgpPeerInformer := bgplbInformerFactory.Bgplb().V1().BgpPeers().Informer()
	bgpEipInformer := bgplbInformerFactory.Bgplb().V1().Eips().Informer()

	controller := &BgpLBController{
		bgpServer:       bgpServer,
		bgpConfInformer: bgpConfInformer,
		bgpPeerInformer: bgpPeerInformer,
		bgpEipInformer:  bgpEipInformer,
	}

	bgpConfInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onBGPConfAdd,
		UpdateFunc: controller.onBGPConfUpdate,
		DeleteFunc: controller.onBGPConfDelete,
	})

	bgpPeerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onBGPPeerAdd,
		UpdateFunc: controller.onBGPPeerUpdate,
		DeleteFunc: controller.onBGPPeerDelete,
	})

	bgpEipInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onBGPEipAdd,
		UpdateFunc: controller.onBGPEipUpdate,
		DeleteFunc: controller.onBGPEipDelete,
	})

	return controller, nil
}

func (controller *BgpLBController) Start() {
	go controller.bgpServer.Serve()

	klog.Info("cache is synced")
}
