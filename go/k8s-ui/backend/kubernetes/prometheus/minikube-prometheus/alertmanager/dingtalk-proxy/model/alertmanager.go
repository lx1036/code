package model

import "time"

// https://prometheus.io/docs/alerting/notifications/#data-structures
// https://prometheus.io/docs/alerting/configuration/#webhook_config
type Notification struct {
	Version           string            `json:"version,omitempty"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver,omitempty"`
	GroupLabels       map[string]string `json:"groupLabels,omitempty"`
	CommonLabels      map[string]string `json:"commonLabels,omitempty"`
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty"`
	ExternalURL       string            `json:"externalURL,omitempty"`
	Alerts            []Alert           `json:"alerts" binding:"required"`
}

type Alert struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}
