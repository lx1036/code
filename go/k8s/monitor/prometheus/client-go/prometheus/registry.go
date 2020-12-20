package prometheus

import (
	"sync"
)

const (
	capDescChan = 10
)

var (
	defaultRegistry              = NewRegistry()
	DefaultRegisterer Registerer = defaultRegistry
	//DefaultGatherer   Gatherer   = defaultRegistry
)

func init() {

}

func MustRegister(cs ...Collector) {
	DefaultRegisterer.MustRegister(cs...)
}

func NewRegistry() *Registry {
	return &Registry{
		collectorsByID:  map[uint64]Collector{},
		descIDs:         map[uint64]struct{}{},
		dimHashesByName: map[string]uint64{},
	}
}

type Collector interface {
	Describe(chan<- *Desc)
	Collect(chan<- Metric)
}

type Registerer interface {
	Register(Collector) error
	MustRegister(...Collector)
	Unregister(Collector) bool
}

type Registry struct {
	mtx                   sync.RWMutex
	collectorsByID        map[uint64]Collector // ID is a hash of the descIDs.
	descIDs               map[uint64]struct{}
	dimHashesByName       map[string]uint64
	uncheckedCollectors   []Collector
	pedanticChecksEnabled bool
}

func (registry *Registry) MustRegister(collectors ...Collector) {
	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			panic(err)
		}
	}
}

func (registry *Registry) Register(collector Collector) error {
	var (
		descChan    = make(chan *Desc, capDescChan)
		newDescIDs  = map[uint64]struct{}{}
		collectorID uint64
	)

	for desc := range descChan {
		if _, exists := registry.descIDs[desc.id]; exists {

		}
		if _, exists := newDescIDs[desc.id]; !exists {
			newDescIDs[desc.id] = struct{}{}
			collectorID += desc.id
		}

	}

	registry.collectorsByID[collectorID] = collector

	return nil
}

func (registry *Registry) Unregister(collector Collector) bool {

	return true
}

/*type Gatherer interface {
	Gather() ([]*metrics.MetricFamily, error)
}
*/

/*func (r Registry) Gather() ([]*interface{}, error) {
	panic("implement me")
}*/

/*type AlreadyRegisteredError struct {
	ExistingCollector,
	NewCollector Collector
}
func (err AlreadyRegisteredError) Error() string {
	return "duplicate metrics collector registration attempted"
}*/
