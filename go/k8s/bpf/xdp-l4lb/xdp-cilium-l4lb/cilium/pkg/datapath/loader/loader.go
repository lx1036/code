package loader

import "sync"

type Loader struct {
	once sync.Once

	// templateCache is the cache of pre-compiled datapaths.
	templateCache *objectCache

	canDisableDwarfRelocations bool
}

// NewLoader returns a new loader.
func NewLoader(canDisableDwarfRelocations bool) *Loader {
	return &Loader{
		canDisableDwarfRelocations: canDisableDwarfRelocations,
	}
}
