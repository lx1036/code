package client

import (
	"context"
	"net/url"

	"github.com/prometheus/common/model"
)

// APIClient is a raw client to the Prometheus Query API.
// It knows how to appropriately deal with generic Prometheus API
// responses, but does not know the specifics of different endpoints.
// You can use this to call query endpoints not represented in Client.
type GenericAPIClient interface {
	// Do makes a request to the Prometheus HTTP API against a particular endpoint.  Query
	// parameters should be in `query`, not `endpoint`.  An error will be returned on HTTP
	// status errors or errors making or unmarshalling the request, as well as when the
	// response has a Status of ResponseError.
	Do(ctx context.Context, verb, endpoint string, query url.Values) (APIResponse, error)
}

// queryClient is a Client that connects to the Prometheus HTTP API.
type queryClient struct {
	api GenericAPIClient
}

func (client *queryClient) Series(ctx context.Context, interval model.Interval, selectors ...Selector) ([]interface{}, error) {
	panic("implement me")
}

func (client *queryClient) Query(ctx context.Context, t model.Time, query Selector) (interface{}, error) {
	panic("implement me")
}

func (client *queryClient) QueryRange(ctx context.Context, r interface{}, query Selector) (interface{}, error) {
	panic("implement me")
}

// NewClientForAPI creates a Client for the given generic Prometheus API client.
func NewClientForAPI(client GenericAPIClient) Client {
	return &queryClient{
		api: client,
	}
}
