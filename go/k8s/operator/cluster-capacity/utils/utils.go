package utils

import (
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
)

func GetMasterFromKubeConfig(kubeConfig string) (string, error) {
	config, err := clientcmd.LoadFromFile(kubeConfig)
	if err != nil {
		return "", fmt.Errorf("can't load kubeconfig file: %v", err)
	}
	context, ok := config.Contexts[config.CurrentContext]
	if !ok {
		return "", fmt.Errorf("failed to get master address from kubeconfig")
	}
	if cluster, ok := config.Clusters[context.Cluster]; ok {
		return cluster.Server, nil
	}
	return "", fmt.Errorf("failed to get master address from kubeconfig")
}

func InitKubeSchedulerConfiguration(options *schedulerOptions.Options) (*schedulerConfig.CompletedConfig, error) {
	config := &schedulerConfig.Config{}
	if err := options.ApplyTo(config); err != nil {
		return nil, fmt.Errorf("unable to get scheduler config: %v", err)
	}
	completedConfig := config.Complete()
	return &completedConfig, nil
}
