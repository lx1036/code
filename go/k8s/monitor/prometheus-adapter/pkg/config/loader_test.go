package config

import (
	"testing"

	"k8s.io/klog/v2"
)

func TestLoaderYamlFile(test *testing.T) {
	config, err := FromFile("config.yaml")
	if err != nil {
		panic(err)
	}

	klog.Infof("rules: %s\n resourceRules: %s\n externalRules: %s",
		config.Rules[0].SeriesQuery,
		config.ResourceRules.CPU.ContainerQuery,
		config.ExternalRules[0].SeriesQuery)
}
