package integration


// IntegrationManager is responsible for management of all integrated applications.
type IntegrationManager interface {
	// IntegrationsGetter is responsible for listing all supported integrations.
	IntegrationsGetter
	// GetState returns state of integration based on its' id.
	GetState(id api.IntegrationID) (*api.IntegrationState, error)
	// Metric returns metric manager that is responsible for management of metric integrations.
	Metric() metric.MetricManager
}
