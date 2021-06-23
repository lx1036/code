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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis/sparkoperator.k9s.io/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SparkApplicationLister helps list SparkApplications.
// All objects returned here must be treated as read-only.
type SparkApplicationLister interface {
	// List lists all SparkApplications in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.SparkApplication, err error)
	// SparkApplications returns an object that can list and get SparkApplications.
	SparkApplications(namespace string) SparkApplicationNamespaceLister
	SparkApplicationListerExpansion
}

// sparkApplicationLister implements the SparkApplicationLister interface.
type sparkApplicationLister struct {
	indexer cache.Indexer
}

// NewSparkApplicationLister returns a new SparkApplicationLister.
func NewSparkApplicationLister(indexer cache.Indexer) SparkApplicationLister {
	return &sparkApplicationLister{indexer: indexer}
}

// List lists all SparkApplications in the indexer.
func (s *sparkApplicationLister) List(selector labels.Selector) (ret []*v1.SparkApplication, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.SparkApplication))
	})
	return ret, err
}

// SparkApplications returns an object that can list and get SparkApplications.
func (s *sparkApplicationLister) SparkApplications(namespace string) SparkApplicationNamespaceLister {
	return sparkApplicationNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SparkApplicationNamespaceLister helps list and get SparkApplications.
// All objects returned here must be treated as read-only.
type SparkApplicationNamespaceLister interface {
	// List lists all SparkApplications in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.SparkApplication, err error)
	// Get retrieves the SparkApplication from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.SparkApplication, error)
	SparkApplicationNamespaceListerExpansion
}

// sparkApplicationNamespaceLister implements the SparkApplicationNamespaceLister
// interface.
type sparkApplicationNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SparkApplications in the indexer for a given namespace.
func (s sparkApplicationNamespaceLister) List(selector labels.Selector) (ret []*v1.SparkApplication, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.SparkApplication))
	})
	return ret, err
}

// Get retrieves the SparkApplication from the indexer for a given namespace and name.
func (s sparkApplicationNamespaceLister) Get(name string) (*v1.SparkApplication, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("sparkapplication"), name)
	}
	return obj.(*v1.SparkApplication), nil
}