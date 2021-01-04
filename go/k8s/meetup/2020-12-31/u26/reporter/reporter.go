package reporter

import (
	"time"
)

type Reporter interface {
	StartRepeatedReport(stopCh <-chan struct{}, start, end time.Time)
}
