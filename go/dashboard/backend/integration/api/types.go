package api

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// IntegrationID is a unique identification string that every integrated app has to provide.
// All ids are kept in this file to minimize the risk of creating conflicts.
type IntegrationID string

// IntegrationState represents integration application state. Provides information about
// health (if dashboard can connect to it) of the integrated application.
// ----------------IMPORTANT----------------
// Until external storage sync is implemented information about state of integration is refreshed
// on every request to ensure that every dashboard replica always returns up-to-date data.
// It does not make dashboard stateful in any way.
// ----------------IMPORTANT----------------
type IntegrationState struct {
	Connected   bool    `json:"connected"`
	LastChecked v1.Time `json:"lastChecked"`
	Error       error   `json:"error"`
}

// Integration represents application integrated into the dashboard. Every application
// has to provide health check and id. Additionally every client supported by integration manager
// has to implement this interface
type Integration interface {
	// HealthCheck is required in order to check state of integration application. We have to
	// be able to connect to it in order to enable it for users. Returns nil if connection
	// can be established, error otherwise.
	HealthCheck() error
	// ID returns unique id of integration application.
	ID() IntegrationID
}
