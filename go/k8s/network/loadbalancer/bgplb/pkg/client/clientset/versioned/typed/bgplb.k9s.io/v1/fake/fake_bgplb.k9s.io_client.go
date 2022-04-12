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

package fake

import (
	v1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/clientset/versioned/typed/bgplb.k9s.io/v1"

	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeBgplbV1 struct {
	*testing.Fake
}

func (c *FakeBgplbV1) BGPPeers() v1.BGPPeerInterface {
	return &FakeBGPPeers{c}
}

func (c *FakeBgplbV1) BgpConves(namespace string) v1.BgpConfInterface {
	return &FakeBgpConves{c, namespace}
}

func (c *FakeBgplbV1) Eips(namespace string) v1.EipInterface {
	return &FakeEips{c, namespace}
}

func (c *FakeBgplbV1) IPPools() v1.IPPoolInterface {
	return &FakeIPPools{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeBgplbV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
