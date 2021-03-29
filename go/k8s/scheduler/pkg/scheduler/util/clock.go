package util

import "time"

// Clock provides an interface for getting the current time
type Clock interface {
	Now() time.Time
}

// RealClock implements a clock using time
type RealClock struct{}

// Now returns the current time with time.Now
func (RealClock) Now() time.Time {
	return time.Now()
}
