/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/apis/ipam.k9s.io/v1"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/client/clientset/versioned/scheme"

	rest "k8s.io/client-go/rest"
)

type IpamV1Interface interface {
	RESTClient() rest.Interface
	IPPoolsGetter
}

// IpamV1Client is used to interact with features provided by the ipam.k9s.io group.
type IpamV1Client struct {
	restClient rest.Interface
}

func (c *IpamV1Client) IPPools() IPPoolInterface {
	return newIPPools(c)
}

// NewForConfig creates a new IpamV1Client for the given config.
func NewForConfig(c *rest.Config) (*IpamV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &IpamV1Client{client}, nil
}

// NewForConfigOrDie creates a new IpamV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *IpamV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new IpamV1Client for the given RESTClient.
func New(c rest.Interface) *IpamV1Client {
	return &IpamV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *IpamV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
