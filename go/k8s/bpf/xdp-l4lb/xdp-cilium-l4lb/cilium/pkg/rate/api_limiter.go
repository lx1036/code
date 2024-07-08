package rate

import (
	"fmt"
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"

	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

// APILimiter is an extension to x/time/rate.Limiter specifically for Cilium
// API calls. It allows to automatically adjust the rate, burst and maximum
// parallel API calls to stay as close as possible to an estimated processing
// time.
type APILimiter struct {
	// name is the name of the API call. This field is immutable after
	// NewAPILimiter()
	name string

	// params is the parameters of the limiter. This field is immutable
	// after NewAPILimiter()
	params APILimiterParameters

	// metrics points to the metrics implementation provided by the caller
	// of the APILimiter. This field is immutable after NewAPILimiter()
	metrics MetricsObserver

	// mutex protects all fields below this line
	mutex lock.RWMutex

	// meanProcessingDuration is the latest mean processing duration,
	// calculated based on processingDurations
	meanProcessingDuration float64

	// processingDurations is the last params.MeanOver processing durations
	processingDurations []time.Duration

	// meanWaitDuration is the latest mean wait duration, calculated based
	// on waitDurations
	meanWaitDuration float64

	// waitDurations is the last params.MeanOver wait durations
	waitDurations []time.Duration

	// parallelRequests is the currently allowed maximum parallel
	// requests. This defaults to params.MaxParallel requests and is then
	// adjusted automatically if params.AutoAdjust is enabled.
	parallelRequests int

	// adjustmentFactor is the latest adjustment factor. It is the ratio
	// between params.EstimatedProcessingDuration and
	// meanProcessingDuration.
	adjustmentFactor float64

	// limiter is the rate limiter based on params.RateLimit and
	// params.RateBurst.
	limiter *rate.Limiter

	// currentRequestsInFlight is the number of parallel API requests
	// currently in flight
	currentRequestsInFlight int

	// requestsProcessed is the total number of processed requests
	requestsProcessed int64

	// requestsScheduled is the total number of scheduled requests
	requestsScheduled int64

	// parallelWaitSemaphore is the semaphore used to implement
	// params.MaxParallel. It is initialized with a capacity of
	// waitSemaphoreResolution and each API request will acquire
	// waitSemaphoreResolution/params.MaxParallel tokens.
	parallelWaitSemaphore *semaphore.Weighted
}

// APILimiterParameters is the configuration of an APILimiter. The structure
// may not be mutated after it has been passed into NewAPILimiter().
type APILimiterParameters struct {
	// EstimatedProcessingDuration is the estimated duration an API call
	// will take. This value is used if AutoAdjust is enabled to
	// automatically adjust rate limits to stay as close as possible to the
	// estimated processing duration.
	EstimatedProcessingDuration time.Duration

	// AutoAdjust enables automatic adjustment of the values
	// ParallelRequests, RateLimit, and RateBurst in order to keep the
	// mean processing duration close to EstimatedProcessingDuration
	AutoAdjust bool

	// MeanOver is the number of entries to keep in order to calculate the
	// mean processing and wait duration
	MeanOver int

	// ParallelRequests is the parallel requests allowed. If AutoAdjust is
	// enabled, the value will adjust automatically.
	ParallelRequests int

	// MaxParallelRequests is the maximum parallel requests allowed. If
	// AutoAdjust is enabled, then the ParalelRequests will never grow
	// above MaxParallelRequests.
	MaxParallelRequests int

	// MinParallelRequests is the minimum parallel requests allowed. If
	// AutoAdjust is enabled, then the ParallelRequests will never fall
	// below MinParallelRequests.
	MinParallelRequests int

	// RateLimit is the initial number of API requests allowed per second.
	// If AutoAdjust is enabled, the value will adjust automatically.
	RateLimit rate.Limit

	// RateBurst is the initial allowed burst of API requests allowed. If
	// AutoAdjust is enabled, the value will adjust automatically.
	RateBurst int

	// MinWaitDuration is the minimum time an API request always has to
	// wait before the Wait() function returns an error.
	MinWaitDuration time.Duration

	// MaxWaitDuration is the maximum time an API request is allowed to
	// wait before the Wait() function returns an error.
	MaxWaitDuration time.Duration

	// Log enables info logging of processed API requests. This should only
	// be used for low frequency API calls.
	Log bool

	// DelayedAdjustmentFactor is percentage of the AdjustmentFactor to be
	// applied to RateBurst and MaxWaitDuration defined as a value between
	// 0.0..1.0. This is used to steer a slower reaction of the RateBurst
	// and ParallelRequests compared to RateLimit.
	DelayedAdjustmentFactor float64

	// SkipInitial is the number of initial API calls for which to not
	// apply any rate limiting. This is useful to define a learning phase
	// in the beginning to allow for auto adjustment before imposing wait
	// durations and rate limiting on API calls.
	SkipInitial int

	// MaxAdjustmentFactor is the maximum adjustment factor when AutoAdjust
	// is enabled. Base values will not adjust more than by this factor.
	MaxAdjustmentFactor float64
}

// NewAPILimiter returns a new APILimiter based on the parameters and metrics implementation
func NewAPILimiter(name string, p APILimiterParameters, metrics MetricsObserver) *APILimiter {
	if p.MeanOver == 0 {
		p.MeanOver = defaultMeanOver
	}

	if p.MinParallelRequests == 0 {
		p.MinParallelRequests = 1
	}

	if p.RateBurst == 0 {
		p.RateBurst = 1
	}

	if p.DelayedAdjustmentFactor == 0.0 {
		p.DelayedAdjustmentFactor = defaultDelayedAdjustmentFactor
	}

	if p.MaxAdjustmentFactor == 0.0 {
		p.MaxAdjustmentFactor = defaultMaxAdjustmentFactor
	}

	l := &APILimiter{
		name:                  name,
		params:                p,
		parallelRequests:      p.ParallelRequests,
		parallelWaitSemaphore: semaphore.NewWeighted(waitSemaphoreResolution),
		metrics:               metrics,
	}

	if p.RateLimit != 0 {
		l.limiter = rate.NewLimiter(p.RateLimit, p.RateBurst)
	}

	return l
}

// APILimiterSet is a set of APILimiter indexed by name
type APILimiterSet struct {
	limiters map[string]*APILimiter
	metrics  MetricsObserver
}

// NewAPILimiterSet creates a new APILimiterSet based on a set of rate limiting
// configurations and the default configuration. Any rate limiter that is
// configured in the config OR the defaults will be configured and made
// available via the Limiter(name) and Wait() function.
func NewAPILimiterSet(config map[string]string, defaults map[string]APILimiterParameters, metrics MetricsObserver) (*APILimiterSet, error) {
	limiters := map[string]*APILimiter{}

	for name, p := range defaults {
		// Merge user config into defaults when provided
		if userConfig, ok := config[name]; ok {
			combinedParams, err := p.MergeUserConfig(userConfig)
			if err != nil {
				return nil, err
			}
			p = combinedParams
		}

		limiters[name] = NewAPILimiter(name, p, metrics)
	}

	for name, c := range config {
		if _, ok := defaults[name]; !ok {
			l, err := NewAPILimiterFromConfig(name, c, metrics)
			if err != nil {
				return nil, fmt.Errorf("unable to parse rate limiting configuration %s=%s: %w", name, c, err)
			}

			limiters[name] = l
		}
	}

	return &APILimiterSet{
		limiters: limiters,
		metrics:  metrics,
	}, nil
}
