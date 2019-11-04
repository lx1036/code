package telemetry

import (
	"log"
	"strings"
)

// AppendUnique adds value to a set named key.
// Set items are unordered. Values in the set
// are unique, but how many times they are
// appended is counted. The value must be
// hashable.
//
// If key is new, a new set will be created for
// values with that key. If key maps to a type
// that is not a counting set, a panic is logged,
// and this is a no-op.
func AppendUnique(key string, value interface{}) {
	if !enabled || isDisabled(key) {
		return
	}
	bufferMu.Lock()
	bufVal, inBuffer := buffer[key]
	setVal, setOk := bufVal.(countingSet)
	if inBuffer && !setOk {
		bufferMu.Unlock()
		log.Printf("[PANIC] Telemetry: key %s already used for non-counting-set value", key)
		return
	}
	if setVal == nil {
		// ensure the buffer is not too full, then add new unique value
		if bufferItemCount >= maxBufferItems {
			bufferMu.Unlock()
			return
		}
		buffer[key] = countingSet{value: 1}
		bufferItemCount++
	} else if setOk {
		// unique value already exists, so just increment counter
		setVal[value]++
	}
	bufferMu.Unlock()
}



// Set puts a value in the buffer to be included
// in the next emission. It overwrites any
// previous value.
//
// This function is safe for multiple goroutines,
// and it is recommended to call this using the
// go keyword after the call to SendHello so it
// doesn't block crucial code.
func Set(key string, val interface{}) {
	if !enabled || isDisabled(key) {
		return
	}
	if _, ok := buffer[key]; !ok {
		if bufferItemCount >= maxBufferItems {
			bufferMu.Unlock()
			return
		}
		bufferItemCount++
	}
	buffer[key] = val
	bufferMu.Unlock()
}

// isDisabled returns whether key is
// a disabled metric key. ALL collection
// functions should call this and not
// save the value if this returns true.
func isDisabled(key string) bool {
	// for keys that are augmented with data, such as
	// "tls_client_hello_ua:<hash>", just
	// check the prefix "tls_client_hello_ua"
	checkKey := key
	if idx := strings.Index(key, ":"); idx > -1 {
		checkKey = key[:idx]
	}

	disabledMetricsMu.RLock()
	_, ok := disabledMetrics[checkKey]
	disabledMetricsMu.RUnlock()
	return ok
}

// StartEmitting sends the current payload and begins the
// transmission cycle for updates. This is the first
// update sent, and future ones will be sent until
// StopEmitting is called.
//
// This function is non-blocking (it spawns a new goroutine).
//
// This function panics if it was called more than once.
// It is a no-op if this package was not initialized.
func StartEmitting() {
	if !enabled {
		return
	}
	updateTimerMu.Lock()
	if updateTimer != nil {
		updateTimerMu.Unlock()
		panic("updates already started")
	}
	updateTimerMu.Unlock()
	updateMu.Lock()
	if updating {
		updateMu.Unlock()
		panic("update already in progress")
	}
	updateMu.Unlock()
	go logEmit(false)
}
