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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	etcdk9siov1 "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/apis/etcd.k9s.io/v1"
	versioned "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/client/clientset/versioned"
	internalinterfaces "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/client/informers/externalversions/internalinterfaces"
	v1 "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/client/listers/etcd.k9s.io/v1"
	time "time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// EtcdBackupInformer provides access to a shared informer and lister for
// EtcdBackups.
type EtcdBackupInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.EtcdBackupLister
}

type etcdBackupInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewEtcdBackupInformer constructs a new informer for EtcdBackup type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewEtcdBackupInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredEtcdBackupInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredEtcdBackupInformer constructs a new informer for EtcdBackup type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredEtcdBackupInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EtcdV1().EtcdBackups(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EtcdV1().EtcdBackups(namespace).Watch(context.TODO(), options)
			},
		},
		&etcdk9siov1.EtcdBackup{},
		resyncPeriod,
		indexers,
	)
}

func (f *etcdBackupInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredEtcdBackupInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *etcdBackupInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&etcdk9siov1.EtcdBackup{}, f.defaultInformer)
}

func (f *etcdBackupInformer) Lister() v1.EtcdBackupLister {
	return v1.NewEtcdBackupLister(f.Informer().GetIndexer())
}