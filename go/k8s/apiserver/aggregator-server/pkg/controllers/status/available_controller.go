package status

import (
	"net/http"

	apiregistrationclient "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/client/clientset/versioned/typed/apiregistration/v1"
	informers "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/client/informers/externalversions/apiregistration/v1"

	"k8s.io/apiserver/pkg/server/egressselector"
	v1informers "k8s.io/client-go/informers/core/v1"
)

// NewAvailableConditionController returns a new AvailableConditionController.
func NewAvailableConditionController(
	apiServiceInformer informers.APIServiceInformer,
	serviceInformer v1informers.ServiceInformer,
	endpointsInformer v1informers.EndpointsInformer,
	apiServiceClient apiregistrationclient.APIServicesGetter,
	proxyTransport *http.Transport,
	proxyCurrentCertKeyContent certKeyFunc,
	serviceResolver ServiceResolver,
	egressSelector *egressselector.EgressSelector,
) (*AvailableConditionController, error) {

}
