package endpoint

import (
    "github.com/sirupsen/logrus"
    "sync/atomic"
)

// getLogger returns a logrus object with EndpointID, containerID and the Endpoint
// revision fields.
func (e *Endpoint) getLogger() *logrus.Entry {
    v := atomic.LoadPointer(&e.logger)
    return (*logrus.Entry)(v)
}
