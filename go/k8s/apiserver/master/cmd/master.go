package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"k8s-lx1036/k8s/apiserver/master"

	"k8s.io/apimachinery/pkg/util/sets"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/filters"
	"k8s.io/apiserver/pkg/server/options"
	clientgoinformers "k8s.io/client-go/informers"
	clientgoclientset "k8s.io/client-go/kubernetes"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "kubeconfig path")
)

func main() {
	flag.Parse()
	if len(*kubeconfig) == 0 {
		klog.Error(fmt.Sprintf("kubeconfig path should not be empty"))
		return
	}

	//kubeClientConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	//if err != nil {
	//	return
	//}
	//clientgoExternalClient, err := clientgoclientset.NewForConfig(kubeClientConfig)
	//if err != nil {
	//	return
	//}
	//versionedInformers := clientgoinformers.NewSharedInformerFactory(clientgoExternalClient, 10*time.Minute)
	kubeAPIServerConfig := CreateKubeAPIServerConfig()
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

// BuildGenericConfig takes the master server options and produces the genericapiserver.Config associated with it
func buildGenericConfig(
	s *options.ServerRunOptions,
	proxyTransport *http.Transport,
) (
	genericConfig *genericapiserver.Config,
	versionedInformers clientgoinformers.SharedInformerFactory,
) {
	genericConfig = genericapiserver.NewConfig(legacyscheme.Codecs)
	genericConfig.MergedResourceConfig = master.DefaultAPIResourceConfigSource()

	genericConfig.LongRunningFunc = filters.BasicLongRunningRequestCheck(
		sets.NewString("watch", "proxy"),
		sets.NewString("attach", "exec", "proxy", "log", "portforward"),
	)

	kubeVersion := version.Get()
	genericConfig.Version = &kubeVersion
	genericConfig.LoopbackClientConfig.ContentConfig.ContentType = "application/vnd.kubernetes.protobuf"
	genericConfig.LoopbackClientConfig.DisableCompression = true
	kubeClientConfig := genericConfig.LoopbackClientConfig
	clientgoExternalClient, err := clientgoclientset.NewForConfig(kubeClientConfig)
	if err != nil {
		return
	}
	versionedInformers = clientgoinformers.NewSharedInformerFactory(clientgoExternalClient, 10*time.Minute)

	return
}

func CreateKubeAPIServerConfig() *master.Config {
	genericConfig, versionedInformers := buildGenericConfig(s.ServerRunOptions, proxyTransport)

	config := &master.Config{
		GenericConfig: genericConfig,
		ExtraConfig: master.ExtraConfig{
			VersionedInformers: versionedInformers,
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
