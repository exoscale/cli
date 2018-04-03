package egoscale

import (
	"context"
	"net/http"
	"time"
)

// Gettable represents an Interface that can be "Get" by the client
type Gettable interface {
	// Get populates the given resource or throws
	Get(context context.Context, client *Client) error
}

// Deletable represents an Interface that can be "Delete" by the client
type Deletable interface {
	// Delete removes the given resource(s) or throws
	Delete(context context.Context, client *Client) error
}

// Listable represents an Interface that can be "List" by the client
type Listable interface {
	// ListRequest builds the list command
	ListRequest() (ListCommand, error)
}

// Client represents the CloudStack API client
type Client struct {
	client    *http.Client
	endpoint  string
	apiKey    string
	apiSecret string
	// PageSize represents the default size for a paginated result
	PageSize int
	// Timeout represents the default timeout for the async requests
	Timeout time.Duration
	// RetryStrategy represents the waiting strategy for polling the async requests
	RetryStrategy RetryStrategyFunc
}

// RetryStrategyFunc represents a how much time to wait between two calls to CloudStack
type RetryStrategyFunc func(int64) time.Duration

// IterateItemFunc represents the callback to iterate a list of results, if false stops
type IterateItemFunc func(interface{}, error) bool
