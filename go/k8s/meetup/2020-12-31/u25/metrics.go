package main

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type Data struct {
	apiName  string
	start    time.Time
	duration time.Duration
}

type Metrics struct {
	Timestamps map[string][]time.Time

	ResponseTimes map[string][]time.Duration
}

func (metrics *Metrics) recordTimestamp(apiName string, start time.Time) {
	metrics.Timestamps[apiName] = append(metrics.Timestamps[apiName], start)
}

func (metrics *Metrics) recordResponseTime(apiName string, duration time.Duration) {
	metrics.ResponseTimes[apiName] = append(metrics.ResponseTimes[apiName], duration)
}

func (metrics *Metrics) startRepeatReport(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			break
		case <-time.Tick(time.Second * 10):
			for apiName, times := range metrics.Timestamps {
				if _, ok := stats[apiName]; !ok {
					stats[apiName] = make(map[string]interface{})
				}

				stats[apiName]["count"] = len(times)
			}

			for apiName, durations := range metrics.ResponseTimes {
				if _, ok := stats[apiName]; !ok {
					stats[apiName] = make(map[string]interface{})
				}

				stats[apiName]["max"] = Max(durations)
				stats[apiName]["min"] = Min(durations)
			}

			for apiName, stat := range stats {
				log.Infof("apiName: %s, max: %v, min: %v", apiName, stat["max"], stat["min"])
			}
		}
	}
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

func NewMetrics(stopCh <-chan struct{}) *Metrics {
	metrics := &Metrics{
		Timestamps:    make(map[string][]time.Time),
		ResponseTimes: make(map[string][]time.Duration),
	}

	go metrics.startRepeatReport(stopCh)

	return metrics
}
