package kube

import (
	"os"

	"github.com/spf13/viper"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetConfigOrDie() *rest.Config {
	var config *rest.Config
	var err error
	if len(viper.GetString("kubeconfig")) != 0 {
		config, err = clientcmd.BuildConfigFromFlags("", viper.GetString("kubeconfig"))
		if err != nil {
			panic(err)
		}

		return config
	}

	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}
	//If file exists so use that config settings
	if _, err = os.Stat(kubeconfigPath); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			panic(err)
		}
	} else { //Use Incluster Configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
	}

	return config
}

// GetClientset gets the clientset for k8s, if ~/.kube/config exists so get that config else incluster config
func GetKubernetesClientset() *kubernetes.Clientset {
	config := GetConfigOrDie()
	return kubernetes.NewForConfigOrDie(config)
}
