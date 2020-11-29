package handler

import (
	"github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/cmd/options"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/metrics"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"sort"
	"strings"
)

const (
	// ConfigmapEnvVarPostfix is a postfix for configmap envVar
	ConfigmapEnvVarPostfix = "CONFIGMAP"
	// SecretEnvVarPostfix is a postfix for secret envVar
	SecretEnvVarPostfix = "SECRET"
)

//Config contains rolling upgrade configuration parameters
type Config struct {
	Namespace           string
	ResourceName        string
	ResourceAnnotations map[string]string
	Annotation          string
	SHAValue            string
	Type                string
}

// GetConfigmapConfig provides utility config for configmap
func GetConfigmapConfig(configmap *corev1.ConfigMap) Config {
	return Config{
		Namespace:           configmap.Namespace,
		ResourceName:        configmap.Name,
		ResourceAnnotations: configmap.Annotations,
		Annotation:          options.ConfigmapUpdateOnChangeAnnotation,
		SHAValue:            GetSHAFromConfigmap(configmap.Data),
		Type:                ConfigmapEnvVarPostfix,
	}
}

// GetSecretConfig provides utility config for secret
func GetSecretConfig(secret *corev1.Secret) Config {
	return Config{
		Namespace:           secret.Namespace,
		ResourceName:        secret.Name,
		ResourceAnnotations: secret.Annotations,
		Annotation:          options.SecretUpdateOnChangeAnnotation,
		SHAValue:            GetSHAFromSecret(secret.Data),
		Type:                SecretEnvVarPostfix,
	}
}

func GetSHAFromConfigmap(data map[string]string) string {
	var values []string
	for k, v := range data {
		values = append(values, k+"="+v)
	}
	sort.Strings(values)
	return util.GenerateSHA(strings.Join(values, ";"))
}

func GetSHAFromSecret(data map[string][]byte) string {
	var values []string
	for k, v := range data {
		values = append(values, k+"="+string(v[:]))
	}
	sort.Strings(values)
	return util.GenerateSHA(strings.Join(values, ";"))
}

// ResourceHandler handles the creation and update of resources
type ResourceHandler interface {
	Handle() error
	GetConfig() (Config, string)
}

// ResourceCreatedHandler contains new objects
type ResourceCreatedHandler struct {
	Resource   interface{}
	Collectors metrics.Collectors
}

func (handler *ResourceCreatedHandler) Handle() error {
	if handler.Resource == nil {
		logrus.Errorf("Resource creation handler received nil resource")
	} else {
		config, _ := handler.GetConfig()
		// process resource based on its type
		doRollingUpgrade(config, handler.Collectors)
	}

	return nil
}
func (handler *ResourceCreatedHandler) GetConfig() (Config, string) {
	var oldSHAData string
	var config Config
	if _, ok := handler.Resource.(*corev1.ConfigMap); ok {
		config = GetConfigmapConfig(handler.Resource.(*corev1.ConfigMap))
	} else if _, ok := handler.Resource.(*corev1.Secret); ok {
		config = GetSecretConfig(handler.Resource.(*corev1.Secret))
	} else {
		logrus.Warnf("Invalid resource: Resource should be 'Secret' or 'Configmap' but found, %v", handler.Resource)
	}
	return config, oldSHAData
}

/*
通过计算configmap.data或secret.data来判断是否更新。
是否更好，为何不通过old/new的(*corev1.ConfigMap).ObjectMeta.ResourceVersion来判断?
*/
// ResourceUpdatedHandler contains updated objects
type ResourceUpdatedHandler struct {
	Resource    interface{}
	OldResource interface{}
	Collectors  metrics.Collectors
}

func (handler *ResourceUpdatedHandler) Handle() error {
	if handler.Resource == nil || handler.OldResource == nil {
		logrus.Errorf("Resource update handler received nil resource")
	} else {
		config, oldSHAData := handler.GetConfig()
		if config.SHAValue != oldSHAData {
			// process resource based on its type
			doRollingUpgrade(config, handler.Collectors)
		}
	}
	return nil
}
func (handler *ResourceUpdatedHandler) GetConfig() (Config, string) {
	var oldSHAData string
	var config Config
	if _, ok := handler.Resource.(*corev1.ConfigMap); ok {
		oldSHAData = GetSHAFromConfigmap(handler.OldResource.(*corev1.ConfigMap).Data)
		config = GetConfigmapConfig(handler.Resource.(*corev1.ConfigMap))
	} else if _, ok := handler.Resource.(*corev1.Secret); ok {
		oldSHAData = GetSHAFromSecret(handler.OldResource.(*corev1.Secret).Data)
		config = GetSecretConfig(handler.Resource.(*corev1.Secret))
	} else {
		logrus.Warnf("Invalid resource: Resource should be 'Secret' or 'Configmap' but found, %v", handler.Resource)
	}
	return config, oldSHAData
}
