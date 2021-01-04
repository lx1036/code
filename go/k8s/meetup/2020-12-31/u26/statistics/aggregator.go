package statistics

import (
	"time"

	"k8s-lx1036/k8s/meetup/2020-12-31/u26/request"
)

type RequestStat struct {
	ApiName string

	Count int

	MaxResponseTime time.Duration
	MinResponseTime time.Duration
	AvgResponseTime time.Duration
}

// 根据原始数据，得到统计数据
type Aggregator struct {
}

func (aggregator *Aggregator) Aggregate(requestInfos []request.RequestInfo) RequestStat {
	count := 0
	var responseTime []time.Duration
	apiName := ""
	for _, requestInfo := range requestInfos {
		count++
		responseTime = append(responseTime, requestInfo.ResponseTime)
		apiName = requestInfo.ApiName
	}

	return RequestStat{
		ApiName:         apiName,
		Count:           count,
		AvgResponseTime: Avg(responseTime),
		MaxResponseTime: Max(responseTime),
		MinResponseTime: Min(responseTime),
	}
}

func Avg(durations []time.Duration) time.Duration {
	sum := int64(0)
	for _, duration := range durations {
		sum += duration.Microseconds()
	}

	t := sum / int64(len(durations))
	return time.Duration(t) * time.Microsecond
}

func Max(durations []time.Duration) time.Duration {
	max := durations[0]
	for _, duration := range durations {
		if duration.Nanoseconds() >= max.Nanoseconds() {
			max = duration
		}
	}

	return max
}

func Min(durations []time.Duration) time.Duration {
	min := durations[0]
	for _, duration := range durations {
		if duration.Nanoseconds() <= min.Nanoseconds() {
			min = duration
		}
	}

	return min
}

func NewAggregator() *Aggregator {
	return new(Aggregator)
}
