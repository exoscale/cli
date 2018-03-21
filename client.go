package egoscale

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"
)

// Get populates the given resource or fails
func (client *Client) Get(g Gettable) error {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.GetWithContext(ctx, g)
}

// GetWithContext populates the given resource or fails
func (client *Client) GetWithContext(ctx context.Context, g Gettable) error {
	return g.Get(ctx, client)
}

// Delete removes the given resource of fails
func (client *Client) Delete(g Deletable) error {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.DeleteWithContext(ctx, g)
}

// DeleteWithContext removes the given resource of fails
func (client *Client) DeleteWithContext(ctx context.Context, g Deletable) error {
	return g.Delete(ctx, client)
}

// List lists the given resource (and paginate till the end)
func (client *Client) List(g Listable) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.ListWithContext(ctx, g)
}

// ListWithContext lists the given resources (and paginate till the end)
func (client *Client) ListWithContext(ctx context.Context, g Listable) ([]interface{}, error) {
	inChan, errChan := client.AsyncListWithContext(ctx, g)

	s := make([]interface{}, 0)
	var err error

	for {
		select {
		case elem, ok := <-inChan:
			if ok {
				s = append(s, elem)
			} else {
				inChan = nil
			}
		case e, ok := <-errChan:
			if ok {
				err = e
			}
			errChan = nil
		case <-ctx.Done():
			err = ctx.Err()
			inChan = nil
			errChan = nil
		}

		if inChan == nil && errChan == nil {
			break
		}
	}

	return s, err
}

// AsyncList lists the given resources and paginates
func (client *Client) AsyncList(g Listable) (<-chan interface{}, <-chan error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.AsyncListWithContext(ctx, g)
}

// AsyncListWithContext lists the given resources and paginates
func (client *Client) AsyncListWithContext(ctx context.Context, g Listable) (<-chan interface{}, <-chan error) {
	return g.List(ctx, client)
}

// NewClientWithTimeout creates a CloudStack API client
//
// Timeout is set to booth the HTTP client and the client itself.
func NewClientWithTimeout(endpoint, apiKey, apiSecret string, timeout time.Duration) *Client {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	cs := &Client{
		client:        client,
		endpoint:      endpoint,
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		PageSize:      50,
		Timeout:       timeout,
		RetryStrategy: FibonacciRetryStrategy,
	}

	return cs
}

// NewClient creates a CloudStack API client with default timeout (60)
func NewClient(endpoint, apiKey, apiSecret string) *Client {
	timeout := time.Duration(60 * time.Second)
	return NewClientWithTimeout(endpoint, apiKey, apiSecret, timeout)
}

// FibonacciRetryStrategy waits for an increasing amount of time following the Fibonacci sequence
func FibonacciRetryStrategy(iteration int64) time.Duration {
	var a, b, i, tmp int64
	a = 0
	b = 1
	for i = 0; i < iteration; i++ {
		tmp = a + b
		a = b
		b = tmp
	}
	return time.Duration(a) * time.Second
}
