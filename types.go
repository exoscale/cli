package egoscale

import (
	"net/http"
	"time"
)

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

// RetryStrategyFunc represents a how much time to wait between two calls to CloudStack
type RetryStrategyFunc func(int64) time.Duration

// Topology represents a view of the servers
type Topology struct {
	Zones          map[string]string
	Images         map[string]map[int64]string
	Profiles       map[string]string
	Keypairs       []string
	SecurityGroups map[string]string
	AffinityGroups map[string]string
}
