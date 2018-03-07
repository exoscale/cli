package egoscale

import "time"

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
