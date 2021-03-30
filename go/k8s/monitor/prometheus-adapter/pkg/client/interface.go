package client

import (
	"context"

	"github.com/prometheus/common/model"
)

// 为何不用 prometheus client_golang? it lacked support for querying the series metadata.

// NB: the official prometheus API client at https://github.com/prometheus/client_golang
// is rather lackluster -- as of the time of writing of this file, it lacked support
// for querying the series metadata, which we need for the adapter. Instead, we use
// this client.

// Selector represents a series selector
type Selector string

// Client is a Prometheus client for the Prometheus HTTP API.
// The "timeout" parameter for the HTTP API is set based on the context's deadline,
// when present and applicable.
type Client interface {
	// Series lists the time series matching the given series selectors
	Series(ctx context.Context, interval model.Interval, selectors ...Selector) ([]Series, error)
	// Query runs a non-range query at the given time.
	Query(ctx context.Context, t model.Time, query Selector) (QueryResult, error)
	// QueryRange runs a range query at the given time.
	QueryRange(ctx context.Context, r Range, query Selector) (QueryResult, error)
}
