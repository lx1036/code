package deployment

import metricapi "k8s-lx1036/dashboard/backend/integration/metric/api"

// DeploymentList contains a list of Deployments in the cluster.
type DeploymentList struct {
	ListMeta          api.ListMeta       `json:"listMeta"`
	CumulativeMetrics []metricapi.Metric `json:"cumulativeMetrics"`

	// Basic information about resources status on the list.
	Status common.ResourceStatus `json:"status"`

	// Unordered list of Deployments.
	Deployments []Deployment `json:"deployments"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// Deployment is a presentation layer view of Kubernetes Deployment resource. This means
// it is Deployment plus additional augmented data we can get from other sources
// (like services that target the same pods).
type Deployment struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// Aggregate information about pods belonging to this Deployment.
	Pods common.PodInfo `json:"pods"`

	// Container images of the Deployment.
	ContainerImages []string `json:"containerImages"`

	// Init Container images of the Deployment.
	InitContainerImages []string `json:"initContainerImages"`
}
