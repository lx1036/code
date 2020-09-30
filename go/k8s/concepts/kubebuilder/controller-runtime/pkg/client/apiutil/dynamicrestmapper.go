package apiutil

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
)

// 运行时自动发现resource type
type dynamicRESTMapper struct {

}

type DynamicRESTMapperOption func(*dynamicRESTMapper) error

func NewDynamicRESTMapper(config *rest.Config, options ...DynamicRESTMapperOption) (meta.RESTMapper, error) {

}
