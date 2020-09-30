package client

import (
	"k8s.io/apimachinery/pkg/conversion/queryparams"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	
	"net/url"
	"errors"
)

type noConversionParamCodec struct{}

func (noConversionParamCodec) EncodeParameters(obj runtime.Object, to schema.GroupVersion) (url.Values, error) {
	return queryparams.Convert(obj)
}

func (noConversionParamCodec) DecodeParameters(parameters url.Values, from schema.GroupVersion, into runtime.Object) error {
	return errors.New("DecodeParameters not implemented on noConversionParamCodec")
}
