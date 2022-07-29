package framework

import (
	"errors"
	"sync"
)

const (
	// NotFound is the not found error message.
	NotFound = "not found"
)

// StateData is a generic type for arbitrary data stored in CycleState.
type StateData interface {
	// Clone is an interface to make a copy of StateData. For performance reasons,
	// clone should make shallow copies for members (e.g., slices or maps) that are not
	// impacted by PreFilter's optional AddPod/RemovePod methods.
	Clone() StateData
}

// StateKey is the type of keys stored in CycleState.
type StateKey string

// CycleState provides a mechanism for plugins to store and retrieve arbitrary data.
// StateData stored by one plugin can be read, altered, or deleted by another plugin.
// CycleState does not provide any data protection, as all plugins are assumed to be
// trusted.
type CycleState struct {
	mx      sync.RWMutex
	storage map[StateKey]StateData
	// if recordPluginMetrics is true, PluginExecutionDuration will be recorded for this cycle.
	recordPluginMetrics bool
}

func (c *CycleState) Write(key StateKey, val StateData) {
	c.storage[key] = val
}

func (c *CycleState) Read(key StateKey) (StateData, error) {
	if v, ok := c.storage[key]; ok {
		return v, nil
	}
	return nil, errors.New(NotFound)
}

// ShouldRecordPluginMetrics returns whether PluginExecutionDuration metrics should be recorded.
func (c *CycleState) ShouldRecordPluginMetrics() bool {
	if c == nil {
		return false
	}
	return c.recordPluginMetrics
}

// SetRecordPluginMetrics sets recordPluginMetrics to the given value.
func (c *CycleState) SetRecordPluginMetrics(flag bool) {
	if c == nil {
		return
	}
	c.recordPluginMetrics = flag
}

// NewCycleState initializes a new CycleState and returns its pointer.
func NewCycleState() *CycleState {
	return &CycleState{
		storage: make(map[StateKey]StateData),
	}
}
