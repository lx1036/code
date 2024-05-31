package endpoint

// unconditionalRLock should be used only for reporting endpoint state
func (e *Endpoint) unconditionalRLock() {
    e.mutex.RLock()
}

// runlock read unlocks endpoint mutex
func (e *Endpoint) runlock() {
    e.mutex.RUnlock()
}

// unconditionalLock should be used only for locking endpoint for
// - setting its state to StateDisconnected
// - handling regular Lock errors
// - reporting endpoint status (like in LogStatus method)
// Use Lock in all other cases
func (e *Endpoint) unconditionalLock() {
    e.mutex.Lock()
}

// Unlock unlocks endpoint mutex
func (e *Endpoint) unlock() {
    e.mutex.Unlock()
}
