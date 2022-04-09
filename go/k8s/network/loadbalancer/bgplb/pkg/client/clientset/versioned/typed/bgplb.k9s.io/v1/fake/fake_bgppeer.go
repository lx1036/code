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
	"context"
	bgplbk9siov1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/apis/bgplb.k9s.io/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeBGPPeers implements BGPPeerInterface
type FakeBGPPeers struct {
	Fake *FakeBgplbV1
}

var bgppeersResource = schema.GroupVersionResource{Group: "bgplb.k9s.io", Version: "v1", Resource: "bgppeers"}

var bgppeersKind = schema.GroupVersionKind{Group: "bgplb.k9s.io", Version: "v1", Kind: "BGPPeer"}

// Get takes name of the bGPPeer, and returns the corresponding bGPPeer object, and an error if there is any.
func (c *FakeBGPPeers) Get(ctx context.Context, name string, options v1.GetOptions) (result *bgplbk9siov1.BGPPeer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(bgppeersResource, name), &bgplbk9siov1.BGPPeer{})
	if obj == nil {
		return nil, err
	}
	return obj.(*bgplbk9siov1.BGPPeer), err
}

// List takes label and field selectors, and returns the list of BGPPeers that match those selectors.
func (c *FakeBGPPeers) List(ctx context.Context, opts v1.ListOptions) (result *bgplbk9siov1.BGPPeerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(bgppeersResource, bgppeersKind, opts), &bgplbk9siov1.BGPPeerList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &bgplbk9siov1.BGPPeerList{ListMeta: obj.(*bgplbk9siov1.BGPPeerList).ListMeta}
	for _, item := range obj.(*bgplbk9siov1.BGPPeerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested bGPPeers.
func (c *FakeBGPPeers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(bgppeersResource, opts))
}

// Create takes the representation of a bGPPeer and creates it.  Returns the server's representation of the bGPPeer, and an error, if there is any.
func (c *FakeBGPPeers) Create(ctx context.Context, bGPPeer *bgplbk9siov1.BGPPeer, opts v1.CreateOptions) (result *bgplbk9siov1.BGPPeer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(bgppeersResource, bGPPeer), &bgplbk9siov1.BGPPeer{})
	if obj == nil {
		return nil, err
	}
	return obj.(*bgplbk9siov1.BGPPeer), err
}

// Update takes the representation of a bGPPeer and updates it. Returns the server's representation of the bGPPeer, and an error, if there is any.
func (c *FakeBGPPeers) Update(ctx context.Context, bGPPeer *bgplbk9siov1.BGPPeer, opts v1.UpdateOptions) (result *bgplbk9siov1.BGPPeer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(bgppeersResource, bGPPeer), &bgplbk9siov1.BGPPeer{})
	if obj == nil {
		return nil, err
	}
	return obj.(*bgplbk9siov1.BGPPeer), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeBGPPeers) UpdateStatus(ctx context.Context, bGPPeer *bgplbk9siov1.BGPPeer, opts v1.UpdateOptions) (*bgplbk9siov1.BGPPeer, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(bgppeersResource, "status", bGPPeer), &bgplbk9siov1.BGPPeer{})
	if obj == nil {
		return nil, err
	}
	return obj.(*bgplbk9siov1.BGPPeer), err
}

// Delete takes name of the bGPPeer and deletes it. Returns an error if one occurs.
func (c *FakeBGPPeers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(bgppeersResource, name), &bgplbk9siov1.BGPPeer{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBGPPeers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(bgppeersResource, listOpts)

	_, err := c.Fake.Invokes(action, &bgplbk9siov1.BGPPeerList{})
	return err
}

// Patch applies the patch and returns the patched bGPPeer.
func (c *FakeBGPPeers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *bgplbk9siov1.BGPPeer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(bgppeersResource, name, pt, data, subresources...), &bgplbk9siov1.BGPPeer{})
	if obj == nil {
		return nil, err
	}
	return obj.(*bgplbk9siov1.BGPPeer), err
}