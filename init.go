package egoscale

import (
	"crypto/tls"
	"net/http"
	"time"
)

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
