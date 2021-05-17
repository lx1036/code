package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"time"

	"k8s-lx1036/k8s/apiserver/master"
	"k8s-lx1036/k8s/apiserver/master/reconcilers"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/filters"
	clientgoinformers "k8s.io/client-go/informers"
	clientgoclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/master/tunneler"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "kubeconfig path")
)

// go run . --kubeconfig=`echo $HOME`/.kube/config
func main() {
	flag.Parse()
	if len(*kubeconfig) == 0 {
		klog.Error(fmt.Sprintf("kubeconfig path should not be empty"))
		return
	}

	nodeTunneler, proxyTransport := CreateNodeDialer()
	kubeAPIServerConfig := CreateKubeAPIServerConfig(nodeTunneler, proxyTransport)
	GenericConfig := &genericapiserver.RecommendedConfig{
		Config:                *kubeAPIServerConfig.GenericConfig,
		SharedInformerFactory: kubeAPIServerConfig.ExtraConfig.VersionedInformers,
	}
	genericServer, err := GenericConfig.Complete().New("apiextensions-apiserver", genericapiserver.NewEmptyDelegate())
	kubeAPIServer, err := CreateKubeAPIServer(kubeAPIServerConfig, genericServer)
	if err != nil {
		return
	}

	err = kubeAPIServer.GenericAPIServer.PrepareRun().Run(genericapiserver.SetupSignalHandler())
	if err != nil {
		return
	}
}

// CreateNodeDialer creates the dialer infrastructure to connect to the nodes.
func CreateNodeDialer() (tunneler.Tunneler, *http.Transport) {
	// Setup nodeTunneler if needed
	var nodeTunneler tunneler.Tunneler
	var proxyDialerFn utilnet.DialFunc
	// Proxying to pods and services is IP-based... don't expect to be able to verify the hostname
	proxyTLSClientConfig := &tls.Config{InsecureSkipVerify: true}
	proxyTransport := utilnet.SetTransportDefaults(&http.Transport{
		DialContext:     proxyDialerFn,
		TLSClientConfig: proxyTLSClientConfig,
	})

	return nodeTunneler, proxyTransport
}

func CreateKubeAPIServerConfig(nodeTunneler tunneler.Tunneler, proxyTransport *http.Transport) *master.Config {
	genericConfig, versionedInformers := buildGenericConfig(proxyTransport)

	config := &master.Config{
		GenericConfig: genericConfig,
		ExtraConfig: master.ExtraConfig{
			VersionedInformers:     versionedInformers,
			EndpointReconcilerType: reconcilers.LeaseEndpointReconcilerType,
		},
	}

	return config
}

// CreateKubeAPIServer creates and wires a workable kube-apiserver
func CreateKubeAPIServer(kubeAPIServerConfig *master.Config,
	delegateAPIServer genericapiserver.DelegationTarget) (*master.Master, error) {
	kubeAPIServer, err := kubeAPIServerConfig.Complete().New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return kubeAPIServer, nil
}

// BuildGenericConfig takes the master server options and produces the genericapiserver.Config associated with it
func buildGenericConfig(proxyTransport *http.Transport) (genericConfig *genericapiserver.Config, versionedInformers clientgoinformers.SharedInformerFactory) {
	genericConfig = genericapiserver.NewConfig(legacyscheme.Codecs)
	genericConfig.MergedResourceConfig = master.DefaultAPIResourceConfigSource()

	genericConfig.LongRunningFunc = filters.BasicLongRunningRequestCheck(
		sets.NewString("watch", "proxy"),
		sets.NewString("attach", "exec", "proxy", "log", "portforward"),
	)

	kubeVersion := version.Get()
	genericConfig.Version = &kubeVersion
	genericConfig.ExternalAddress = "127.0.0.1:8082"

	kubeClientConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return
	}
	genericConfig.LoopbackClientConfig = kubeClientConfig
	genericConfig.LoopbackClientConfig.ContentConfig.ContentType = "application/vnd.kubernetes.protobuf"
	genericConfig.LoopbackClientConfig.DisableCompression = true
	clientgoExternalClient, err := clientgoclientset.NewForConfig(genericConfig.LoopbackClientConfig)
	if err != nil {
		return
	}

	versionedInformers = clientgoinformers.NewSharedInformerFactory(clientgoExternalClient, 10*time.Minute)

	return
}
