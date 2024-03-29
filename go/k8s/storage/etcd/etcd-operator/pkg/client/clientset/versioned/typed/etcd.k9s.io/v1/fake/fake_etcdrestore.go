/*
Copyright 2021 The Kubernetes Authors.

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
	etcdk9siov1 "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/apis/etcd.k9s.io/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeEtcdRestores implements EtcdRestoreInterface
type FakeEtcdRestores struct {
	Fake *FakeEtcdV1
	ns   string
}

var etcdrestoresResource = schema.GroupVersionResource{Group: "etcd.k9s.io", Version: "v1", Resource: "etcdrestores"}

var etcdrestoresKind = schema.GroupVersionKind{Group: "etcd.k9s.io", Version: "v1", Kind: "EtcdRestore"}

// Get takes name of the etcdRestore, and returns the corresponding etcdRestore object, and an error if there is any.
func (c *FakeEtcdRestores) Get(ctx context.Context, name string, options v1.GetOptions) (result *etcdk9siov1.EtcdRestore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(etcdrestoresResource, c.ns, name), &etcdk9siov1.EtcdRestore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*etcdk9siov1.EtcdRestore), err
}

// List takes label and field selectors, and returns the list of EtcdRestores that match those selectors.
func (c *FakeEtcdRestores) List(ctx context.Context, opts v1.ListOptions) (result *etcdk9siov1.EtcdRestoreList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(etcdrestoresResource, etcdrestoresKind, c.ns, opts), &etcdk9siov1.EtcdRestoreList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &etcdk9siov1.EtcdRestoreList{ListMeta: obj.(*etcdk9siov1.EtcdRestoreList).ListMeta}
	for _, item := range obj.(*etcdk9siov1.EtcdRestoreList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested etcdRestores.
func (c *FakeEtcdRestores) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(etcdrestoresResource, c.ns, opts))

}

// Create takes the representation of a etcdRestore and creates it.  Returns the server's representation of the etcdRestore, and an error, if there is any.
func (c *FakeEtcdRestores) Create(ctx context.Context, etcdRestore *etcdk9siov1.EtcdRestore, opts v1.CreateOptions) (result *etcdk9siov1.EtcdRestore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(etcdrestoresResource, c.ns, etcdRestore), &etcdk9siov1.EtcdRestore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*etcdk9siov1.EtcdRestore), err
}

// Update takes the representation of a etcdRestore and updates it. Returns the server's representation of the etcdRestore, and an error, if there is any.
func (c *FakeEtcdRestores) Update(ctx context.Context, etcdRestore *etcdk9siov1.EtcdRestore, opts v1.UpdateOptions) (result *etcdk9siov1.EtcdRestore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(etcdrestoresResource, c.ns, etcdRestore), &etcdk9siov1.EtcdRestore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*etcdk9siov1.EtcdRestore), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeEtcdRestores) UpdateStatus(ctx context.Context, etcdRestore *etcdk9siov1.EtcdRestore, opts v1.UpdateOptions) (*etcdk9siov1.EtcdRestore, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(etcdrestoresResource, "status", c.ns, etcdRestore), &etcdk9siov1.EtcdRestore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*etcdk9siov1.EtcdRestore), err
}

// Delete takes name of the etcdRestore and deletes it. Returns an error if one occurs.
func (c *FakeEtcdRestores) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(etcdrestoresResource, c.ns, name), &etcdk9siov1.EtcdRestore{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeEtcdRestores) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(etcdrestoresResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &etcdk9siov1.EtcdRestoreList{})
	return err
}

// Patch applies the patch and returns the patched etcdRestore.
func (c *FakeEtcdRestores) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *etcdk9siov1.EtcdRestore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(etcdrestoresResource, c.ns, name, pt, data, subresources...), &etcdk9siov1.EtcdRestore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*etcdk9siov1.EtcdRestore), err
}
