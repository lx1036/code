package request

import "time"

type RequestInfo struct {
	ApiName string

	ResponseTime time.Duration

	Timestamp time.Time
}
