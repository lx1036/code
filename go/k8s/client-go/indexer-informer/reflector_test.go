package indexer_informer

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"testing"
)

type testLW struct {
	ListFunc  func(options metav1.ListOptions) (runtime.Object, error)
	WatchFunc func(options metav1.ListOptions) (watch.Interface, error)
}

func (t *testLW) List(options metav1.ListOptions) (runtime.Object, error) {
	return t.ListFunc(options)
}
func (t *testLW) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return t.WatchFunc(options)
}

func TestCloseWatchChannelOnError(test *testing.T) {
	//reflector := cache.NewReflector(&testLW{}, &v1.Pod{}, cache.NewStore(cache.MetaNamespaceKeyFunc), 0)
	//pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name:"bar"}}
	//fakeWatch := watch.NewFake()
	//reflector.
}
