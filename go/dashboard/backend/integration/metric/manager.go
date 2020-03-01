package metric

import (
	integrationapi "k8s-lx1036/dashboard/backend/integration/api"
	metricapi "k8s-lx1036/dashboard/backend/integration/metric/api"
	"time"
)

// MetricManager is responsible for management of all integrated applications related to metrics.
type MetricManager interface {
	// AddClient adds metric client to client list supported by this manager.
	AddClient(metricapi.MetricClient) MetricManager
	// Client returns active Metric client.
	Client() metricapi.MetricClient
	// Enable is responsible for switching active client if given integration application id
	// is found and related application is healthy (we can connect to it).
	Enable(integrationapi.IntegrationID) error
	// EnableWithRetry works similar to enable. It runs in a separate thread and tries to enable integration with given
	// id every 'period' seconds.
	EnableWithRetry(id integrationapi.IntegrationID, period time.Duration)
	// List returns list of available metric related integrations.
	List() []integrationapi.Integration
	// ConfigureSidecar configures and adds sidecar to clients list.
	ConfigureSidecar(host string) MetricManager
	// ConfigureHeapster configures and adds sidecar to clients list.
	ConfigureHeapster(host string) MetricManager
}
