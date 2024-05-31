package endpoint

// unconditionalRLock should be used only for reporting endpoint state
func (e *Endpoint) unconditionalRLock() {
    e.mutex.RLock()
}

// runlock read unlocks endpoint mutex
func (e *Endpoint) runlock() {
    e.mutex.RUnlock()
}
