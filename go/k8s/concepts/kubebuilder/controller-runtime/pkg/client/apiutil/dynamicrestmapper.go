package apiutil

import (
	"errors"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sync"
	"time"
)

var (
	// defaultRefilRate is the default rate at which potential calls are
	// added back to the "bucket" of allowed calls.
	defaultRefillRate = 5
	// defaultLimitSize is the default starting/max number of potential calls
	// per second.  Once a call is used, it's added back to the bucket at a rate
	// of defaultRefillRate per second.
	defaultLimitSize = 5
)

// dynamicLimiter holds a rate limiter used to throttle chatty RESTMapper users.
type dynamicLimiter struct {
	*rate.Limiter
}

// checkRate returns an ErrRateLimited if too many API calls have been made
// within the set limit.
func (b *dynamicLimiter) checkRate() error {
	res := b.Reserve()
	if res.Delay() == 0 {
		return nil
	}
	res.Cancel()
	return ErrRateLimited{res.Delay()}
}

// ErrRateLimited is returned by a RESTMapper method if the number of API
// calls has exceeded a limit within a certain time period.
type ErrRateLimited struct {
	// Duration to wait until the next API call can be made.
	Delay time.Duration
}

func (e ErrRateLimited) Error() string {
	return "too many API calls to the RESTMapper within a timeframe"
}

// 运行时自动发现resource type
type dynamicRESTMapper struct {
	mu           sync.RWMutex // protects the following fields
	staticMapper meta.RESTMapper
	limiter      *dynamicLimiter
	newMapper    func() (meta.RESTMapper, error)

	lazy bool
	// Used for lazy init.
	initOnce sync.Once
}

func (drm *dynamicRESTMapper) KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	panic("implement me")
}

func (drm *dynamicRESTMapper) KindsFor(resource schema.GroupVersionResource) ([]schema.GroupVersionKind, error) {
	panic("implement me")
}

func (drm *dynamicRESTMapper) ResourceFor(input schema.GroupVersionResource) (schema.GroupVersionResource, error) {
	panic("implement me")
}

func (drm *dynamicRESTMapper) ResourcesFor(input schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	panic("implement me")
}

// init initializes drm only once if drm is lazy.
func (drm *dynamicRESTMapper) init() (err error) {
	drm.initOnce.Do(func() {
		if drm.lazy {
			err = drm.setStaticMapper()
		}
	})
	return err
}

// checkAndReload attempts to call the given callback, which is assumed to be dependent
// on the data in the restmapper.
//
// If the callback returns a NoKindMatchError, it will attempt to reload
// the RESTMapper's data and re-call the callback once that's occurred.
// If the callback returns any other error, the function will return immediately regardless.
//
// It will take care
// ensuring that reloads are rate-limitted and that extraneous calls aren't made.
// It's thread-safe, and worries about thread-safety for the callback (so the callback does
// not need to attempt to lock the restmapper).
func (drm *dynamicRESTMapper) checkAndReload(needsReloadErr error, checkNeedsReload func() error) error {
	// first, check the common path -- data is fresh enough
	// (use an IIFE for the lock's defer)
	err := func() error {
		drm.mu.RLock()
		defer drm.mu.RUnlock()

		return checkNeedsReload()
	}()

	// NB(directxman12): `Is` and `As` have a confusing relationship --
	// `Is` is like `== or does this implement .Is`, whereas `As` says
	// `can I type-assert into`
	needsReload := errors.As(err, &needsReloadErr)
	if !needsReload {
		return err
	}

	// if the data wasn't fresh, we'll need to try and update it, so grab the lock...
	drm.mu.Lock()
	defer drm.mu.Unlock()

	// ... and double-check that we didn't reload in the meantime
	err = checkNeedsReload()
	needsReload = errors.As(err, &needsReloadErr)
	if !needsReload {
		return err
	}

	// we're still stale, so grab a rate-limit token if we can...
	if err := drm.limiter.checkRate(); err != nil {
		return err
	}

	// ...reload...
	if err := drm.setStaticMapper(); err != nil {
		return err
	}

	// ...and return the results of the closure regardless
	return checkNeedsReload()
}
func (drm *dynamicRESTMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	if err := drm.init(); err != nil {
		return nil, err
	}
	var mapping *meta.RESTMapping
	err := drm.checkAndReload(&meta.NoKindMatchError{}, func() error {
		var err error
		mapping, err = drm.staticMapper.RESTMapping(gk, versions...)
		return err
	})
	return mapping, err
}
func (drm *dynamicRESTMapper) RESTMappings(gk schema.GroupKind, versions ...string) ([]*meta.RESTMapping, error) {
	if err := drm.init(); err != nil {
		return nil, err
	}
	var mappings []*meta.RESTMapping
	err := drm.checkAndReload(&meta.NoKindMatchError{}, func() error {
		var err error
		mappings, err = drm.staticMapper.RESTMappings(gk, versions...)
		return err
	})
	return mappings, err
}

func (drm *dynamicRESTMapper) ResourceSingularizer(resource string) (singular string, err error) {
	panic("implement me")
}

type DynamicRESTMapperOption func(*dynamicRESTMapper) error

// setStaticMapper sets drm's staticMapper by querying its client, regardless
// of reload backoff.
func (drm *dynamicRESTMapper) setStaticMapper() error {
	newMapper, err := drm.newMapper()
	if err != nil {
		return err
	}
	drm.staticMapper = newMapper
	return nil
}

// NewDynamicRESTMapper returns a dynamic RESTMapper for cfg. The dynamic
// RESTMapper dynamically discovers resource types at runtime. opts
// configure the RESTMapper.
func NewDynamicRESTMapper(config *rest.Config, options ...DynamicRESTMapperOption) (meta.RESTMapper, error) {
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	drm := &dynamicRESTMapper{
		limiter: &dynamicLimiter{
			rate.NewLimiter(rate.Limit(defaultRefillRate), defaultLimitSize),
		},
		newMapper: func() (meta.RESTMapper, error) {
			groupResources, err := restmapper.GetAPIGroupResources(client)
			if err != nil {
				return nil, err
			}
			return restmapper.NewDiscoveryRESTMapper(groupResources), nil
		},
	}

	for _, opt := range options {
		if err = opt(drm); err != nil {
			return nil, err
		}
	}
	if !drm.lazy {
		if err := drm.setStaticMapper(); err != nil {
			return nil, err
		}
	}

	return drm, nil
}
