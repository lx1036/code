package informer

import (
	"k8s.io/client-go/tools/cache"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
)

type ConvertFunc func(obj interface{}) interface{}

// NewInformer @see k8s.io/client-go/tools/cache/controller.go::NewInformer, 加了 convertFunc 用来转换成自定义的对象
func NewInformer(
	lw cache.ListerWatcher,
	objType runtime.Object,
	resyncPeriod time.Duration,
	h cache.ResourceEventHandler,
	convertFunc ConvertFunc,
) (cache.Store, cache.Controller) {
	// This will hold the client state, as we know it.
	clientState := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)

	return clientState, NewInformerWithStore(lw, objType, resyncPeriod, h, clientState, convertFunc)
}

func NewInformerWithStore(
	lw cache.ListerWatcher,
	objType runtime.Object,
	resyncPeriod time.Duration,
	h cache.ResourceEventHandler,
	clientState cache.Store,
	convertFunc ConvertFunc,
) cache.Controller {
	fifo := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KeyFunction:           cache.MetaNamespaceKeyFunc,
		KnownObjects:          clientState,
		EmitDeltaTypeReplaced: true,
	})

	cfg := &cache.Config{
		Queue:            fifo,
		ListerWatcher:    lw,
		ObjectType:       objType,
		FullResyncPeriod: resyncPeriod,
		RetryOnError:     false,

		Process: func(obj interface{}) error {
			// from oldest to newest
			for _, d := range obj.(cache.Deltas) {
				obj := d.Object
				if transformer != nil {
					var err error
					obj, err = transformer(obj)
					if err != nil {
						return err
					}
				}

				switch d.Type {
				case Sync, Replaced, Added, Updated:
					if old, exists, err := clientState.Get(obj); err == nil && exists {
						if err := clientState.Update(obj); err != nil {
							return err
						}
						h.OnUpdate(old, obj)
					} else {
						if err := clientState.Add(obj); err != nil {
							return err
						}
						h.OnAdd(obj)
					}
				case Deleted:
					if err := clientState.Delete(obj); err != nil {
						return err
					}
					h.OnDelete(obj)
				}
			}
			return nil
		},
	}

	return cache.New(cfg)
}
