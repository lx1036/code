package client

import "k8s.io/apimachinery/pkg/runtime"

type typedClient struct {
	cache      *clientCache
	paramCodec runtime.ParameterCodec
}
