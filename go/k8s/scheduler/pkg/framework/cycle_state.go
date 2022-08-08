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
	storage sync.Map

	// if recordPluginMetrics is true, PluginExecutionDuration will be recorded for this cycle.
	recordPluginMetrics bool
}

func NewCycleState() *CycleState {
	return &CycleState{}
}

func (c *CycleState) Write(key StateKey, val StateData) {
	c.storage.Store(key, val)
}

func (c *CycleState) Read(key StateKey) (StateData, error) {
	if v, ok := c.storage.Load(key); ok {
		return v.(StateData), nil
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

func (c *CycleState) Clone() *CycleState {
	if c == nil {
		return nil
	}
	copyState := NewCycleState()
	c.storage.Range(func(k, v interface{}) bool {
		copyState.storage.Store(k, v.(StateData).Clone())
		return true
	})

	return copyState
}
