package v3

import "sync"

// Credentials store holds Exoscale API credentials: API key & API secret.
// Structure is safe for concurrent use.
type Credentials struct {
	apiKey    string
	apiSecret string

	mx sync.RWMutex
}

// NewCredentials creates a new API Credentials store using provided API key & secret.
func NewCredentials(apiKey, apiSecret string) *Credentials {
	return &Credentials{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

// Update updates existing Credentials store with new API key & secret.
// Function locks the store to prevent reads.
func (c *Credentials) Update(apiKey, apiSecret string) {
	c.mx.Lock()
	c.apiKey = apiKey
	c.apiSecret = apiSecret
	c.mx.Unlock()
}

// APIKey returns API key from Credentials store.
// Function will prevent any writes, but allow other reads on Credentials store.
func (c *Credentials) APIKey() string {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.apiKey
}

// APIKey returns API secret from Credentials store.
// Function will prevent any writes, but allow other reads on Credentials store.
func (c *Credentials) APISecret() string {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.apiSecret
}
