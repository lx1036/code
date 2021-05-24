package options

import (
	"fmt"
	"io/ioutil"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"
)

func loadConfigFromFile(file string) (*config.KubeSchedulerConfiguration, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return loadConfig(data)
}

func loadConfig(data []byte) (*config.KubeSchedulerConfiguration, error) {
	// The UniversalDecoder runs defaulting and returns the internal type by default.
	obj, gvk, err := config.Codecs.UniversalDecoder().Decode(data, nil, nil)
	if err != nil {
		return nil, err
	}
	if cfgObj, ok := obj.(*config.KubeSchedulerConfiguration); ok {
		return cfgObj, nil
	}
	return nil, fmt.Errorf("couldn't decode as KubeSchedulerConfiguration, got %s: ", gvk)
}
