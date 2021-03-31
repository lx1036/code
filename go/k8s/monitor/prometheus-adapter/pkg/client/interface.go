package client

import (
	"context"
	"time"

	"github.com/prometheus/common/model"
)

// 为何不用 prometheus client_golang? it lacked support for querying the series metadata.

// NB: the official prometheus API client at https://github.com/prometheus/client_golang
// is rather lackluster -- as of the time of writing of this file, it lacked support
// for querying the series metadata, which we need for the adapter. Instead, we use
// this client.

// Selector represents a series selector
type Selector string

// Range represents a sliced time range with increments.
type Range struct {
	// Start and End are the boundaries of the time range.
	Start, End model.Time
	// Step is the maximum time between two slices within the boundaries.
	Step time.Duration
}

// QueryResult is the result of a query.
// Type will always be set, as well as one of the other fields, matching the type.
type QueryResult struct {
	Type model.ValueType

	Vector *model.Vector
	Scalar *model.Scalar
	Matrix *model.Matrix
}

// Client is a Prometheus client for the Prometheus HTTP API.
// The "timeout" parameter for the HTTP API is set based on the context's deadline,
// when present and applicable.
type Client interface {
	// Series lists the time series matching the given series selectors
	Series(ctx context.Context, interval model.Interval, selectors ...Selector) ([]model.LabelSet, error)
	// Query runs a non-range query at the given time.
	Query(ctx context.Context, t model.Time, query Selector) (QueryResult, error)
	// QueryRange runs a range query at the given time.
	QueryRange(ctx context.Context, r Range, query Selector) (QueryResult, error)
}
