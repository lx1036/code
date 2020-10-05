package client

import "k8s.io/apimachinery/pkg/runtime"

type unstructuredClient struct {
	cache      *clientCache
	paramCodec runtime.ParameterCodec
}
