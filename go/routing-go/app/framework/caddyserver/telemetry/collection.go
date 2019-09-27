package telemetry

import "log"

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






