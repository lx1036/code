package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus/common/model"
)

// ResponseStatus is the type of response from the API: succeeded or error.
type ResponseStatus string

// ErrorType is the type of the API error.
type ErrorType string

const (
	ErrBadData     ErrorType = "bad_data"
	ErrTimeout               = "timeout"
	ErrCanceled              = "canceled"
	ErrExec                  = "execution"
	ErrBadResponse           = "bad_response"
)

// Error is an error returned by the API.
type Error struct {
	Type ErrorType
	Msg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Msg)
}

// APIResponse represents the raw response returned by the API.
type APIResponse struct {
	// Status indicates whether this request was successful or whether it errored out.
	Status ResponseStatus `json:"status"`
	// Data contains the raw data response for this request.
	Data json.RawMessage `json:"data"`

	// ErrorType is the type of error, if this is an error response.
	ErrorType ErrorType `json:"errorType"`
	// Error is the error message, if this is an error response.
	Error string `json:"error"`
}

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

// httpAPIClient is a GenericAPIClient implemented in terms of an underlying http.Client.
type httpAPIClient struct {
	client  *http.Client
	baseURL *url.URL
}

func (c *httpAPIClient) Do(ctx context.Context, verb, endpoint string, query url.Values) (APIResponse, error) {
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, endpoint)
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(verb, u.String(), nil)
	if err != nil {
		return APIResponse{}, fmt.Errorf("error constructing HTTP request to Prometheus: %v", err)
	}
	req.WithContext(ctx)

	resp, err := c.client.Do(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return APIResponse{}, err
	}
	//var body io.Reader = resp.Body
	var res APIResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return APIResponse{}, &Error{
			Type: ErrBadResponse,
			Msg:  err.Error(),
		}
	}

	return res, nil
}

// NewGenericAPIClient builds a new generic Prometheus API client for the given base URL and HTTP Client.
func NewGenericAPIClient(client *http.Client, baseURL *url.URL) GenericAPIClient {
	return &httpAPIClient{
		client:  client,
		baseURL: baseURL,
	}
}

const (
	queryURL      = "/api/v1/query"
	queryRangeURL = "/api/v1/query_range"
	seriesURL     = "/api/v1/series"
)

// queryClient is a Client that connects to the Prometheus HTTP API.
type queryClient struct {
	api GenericAPIClient
}

// Series represents a description of a series: a name and a set of labels.
// Series is roughly equivalent to model.Metrics, but has easy access to name
// and the set of non-name labels.
type Series struct {
	Name   string
	Labels model.LabelSet
}

func (client *queryClient) Series(ctx context.Context, interval model.Interval, selectors ...Selector) ([]model.LabelSet, error) {
	vals := url.Values{}
	if interval.Start != 0 {
		vals.Set("start", interval.Start.String())
	}
	if interval.End != 0 {
		vals.Set("end", interval.End.String())
	}

	for _, selector := range selectors {
		vals.Add("match[]", string(selector))
	}
	res, err := client.api.Do(ctx, "GET", seriesURL, vals)
	if err != nil {
		return nil, err
	}

	var seriesRes []model.LabelSet
	err = json.Unmarshal(res.Data, &seriesRes)

	return seriesRes, err
}

func (client *queryClient) Query(ctx context.Context, t model.Time, query Selector) (QueryResult, error) {
	panic("implement me")
}

func (client *queryClient) QueryRange(ctx context.Context, r Range, query Selector) (QueryResult, error) {
	panic("implement me")
}

// NewClientForAPI creates a Client for the given generic Prometheus API client.
func NewClientForAPI(client GenericAPIClient) Client {
	return &queryClient{
		api: client,
	}
}
