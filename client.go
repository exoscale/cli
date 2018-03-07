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
	// Delete removes the given resources or throws
	Delete(context context.Context, client *Client) error
}

// Client represents the CloudStack API client
type Client struct {
	client    *http.Client
	endpoint  string
	apiKey    string
	apiSecret string
	// Timeout represents the default timeout for the async requests
	Timeout time.Duration
	// RetryStrategy represents the waiting strategy for polling the async requests
	RetryStrategy RetryStrategyFunc
}

// Get populates the given resource or fails
func (client *Client) Get(g Gettable) error {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return g.Get(ctx, client)
}

// GetWithContext populates the given resource or fails
func (client *Client) GetWithContext(ctx context.Context, g Gettable) error {
	return g.Get(ctx, client)
}

// Delete removes the given resource of fails
func (client *Client) Delete(g Deletable) error {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return g.Delete(ctx, client)
}

// DeleteWithContext removes the given resource of fails
func (client *Client) DeleteWithContext(ctx context.Context, g Deletable) error {
	return g.Delete(ctx, client)
}
