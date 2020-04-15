package kubernetes

import (
	kube_rest "k8s.io/client-go/rest"
	"net/url"
)

const (
	defaultInClusterConfig = true
)

func GetKubeClientConfig(uri *url.URL) (*kube_rest.Config, error) {
	var kubeConfig *kube_rest.Config
	var err error
	inClusterConfig := defaultInClusterConfig
	if inClusterConfig {
		kubeConfig, err = kube_rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	return kubeConfig, nil
}
