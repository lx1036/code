package statistics

import "time"

type RequestInfo struct {
	ApiName string
	
	ResponseTime time.Duration
}

