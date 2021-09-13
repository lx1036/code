package etcd3

import (
	"k8s-lx1036/k8s/apiserver/pkg/storage"
	"k8s-lx1036/k8s/apiserver/pkg/storage/value"

	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/apimachinery/pkg/runtime"
)

type watcher struct {
	client      *clientv3.Client
	codec       runtime.Codec
	versioner   storage.Versioner
	transformer value.Transformer
}

func newWatcher(client *clientv3.Client, codec runtime.Codec, versioner storage.Versioner,
	transformer value.Transformer) *watcher {
	return &watcher{
		client:      client,
		codec:       codec,
		versioner:   versioner,
		transformer: transformer,
	}
}
