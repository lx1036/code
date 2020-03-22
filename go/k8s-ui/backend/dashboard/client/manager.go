package client

import (
	pluginclientset "k8s-lx1036/k8s-ui/backend/dashboard/controllers/plugin/client/clientset/versioned"
	"k8s-lx1036/k8s-ui/backend/dashboard/mode"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

var DefaultClientManager *ClientManager

type ClientManager struct {
	inClusterConfig *rest.Config

	kubeConfigPath string
	// protocol://address:port
	apiServerHost string

	//
	insecureConfig *rest.Config

	// k8s client created without providing auth info
	insecureClient kubernetes.Interface

	// Plugin client created without providing auth info. It uses permissions granted to
	// service account used by dashboard or kubeconfig file if it was passed during dashboard init.
	insecurePluginClient pluginclientset.Interface
}

func (manager *ClientManager) initInClusterConfig() {
	if len(manager.kubeConfigPath) > 0 || len(manager.apiServerHost) > 0 {
		log.Printf("skip in-cluster config")
		return
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	manager.inClusterConfig = config
}

func (manager *ClientManager) initInsecureClients() {
	config, err := clientcmd.BuildConfigFromFlags(manager.apiServerHost, manager.kubeConfigPath)
	if err != nil {
		panic(err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	pluginClient, err := pluginclientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	manager.insecureConfig = config
	manager.insecureClient = k8sClient
	manager.insecurePluginClient = pluginClient
}

func (manager *ClientManager) GetPluginClient() pluginclientset.Interface {
	return manager.insecurePluginClient
}

func (manager *ClientManager) Client() kubernetes.Interface {
	if mode.Mode() == mode.TestMode {
		clientSet := fake.NewSimpleClientset()
		// TODO: add something
		return clientSet
	}

	return manager.insecureClient
}

func NewClientManager(kubeConfigPath, apiServerHost string) *ClientManager {
	manager := &ClientManager{
		kubeConfigPath: kubeConfigPath,
		apiServerHost:  apiServerHost,
	}

	manager.initInClusterConfig()
	manager.initInsecureClients()

	return manager
}
