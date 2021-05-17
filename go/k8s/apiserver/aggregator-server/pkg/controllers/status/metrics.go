package status

import (
	"sync"

	"k8s.io/component-base/metrics"
)

type availabilityMetrics struct {
	unavailableCounter *metrics.CounterVec

	*availabilityCollector
}

func newAvailabilityMetrics() *availabilityMetrics {
	return &availabilityMetrics{
		unavailableCounter: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Name:           "aggregator_unavailable_apiservice_total",
				Help:           "Counter of APIServices which are marked as unavailable broken down by APIService name and reason.",
				StabilityLevel: metrics.ALPHA,
			},
			[]string{"name", "reason"},
		),
		availabilityCollector: newAvailabilityCollector(),
	}
}

// Register registers apiservice availability metrics.
func (m *availabilityMetrics) Register(registrationFunc func(metrics.Registerable) error,
	customRegistrationFunc func(metrics.StableCollector) error) error {
	err := registrationFunc(m.unavailableCounter)
	if err != nil {
		return err
	}

	err = customRegistrationFunc(m.availabilityCollector)
	if err != nil {
		return err
	}

	return nil
}

// UnavailableCounter returns a counter to track apiservices marked as unavailable.
func (m *availabilityMetrics) UnavailableCounter(apiServiceName, reason string) metrics.CounterMetric {
	return m.unavailableCounter.WithLabelValues(apiServiceName, reason)
}

type availabilityCollector struct {
	metrics.BaseStableCollector

	mtx            sync.RWMutex
	availabilities map[string]bool
}

func newAvailabilityCollector() *availabilityCollector {
	return &availabilityCollector{
		availabilities: make(map[string]bool),
	}
}

// ForgetAPIService removes the availability gauge of the given apiservice.
func (c *availabilityCollector) ForgetAPIService(apiServiceKey string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	delete(c.availabilities, apiServiceKey)
}

// SetAPIServiceAvailable sets the given apiservice availability gauge to available.
func (c *availabilityCollector) SetAPIServiceAvailable(apiServiceKey string) {
	c.setAPIServiceAvailability(apiServiceKey, true)
}

func (c *availabilityCollector) setAPIServiceAvailability(apiServiceKey string, availability bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.availabilities[apiServiceKey] = availability
}

// SetAPIServiceUnavailable sets the given apiservice availability gauge to unavailable.
func (c *availabilityCollector) SetAPIServiceUnavailable(apiServiceKey string) {
	c.setAPIServiceAvailability(apiServiceKey, false)
}
