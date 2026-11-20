package credentials

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrMissingIncomplete = errors.New("missing or incomplete API credentials")
)

type Value struct {
	APIKey    string
	APISecret string
}

// IsSet returns true if the credentials Value has both APIKey and APISecret.
func (v Value) IsSet() bool {
	return v.APIKey != "" && v.APISecret != ""
}

type Provider interface {
	// Retrieve returns nil if it successfully retrieved the value.
	// Error is returned if the value were not obtainable, or empty.
	Retrieve() (Value, error)

	// IsExpired returns if the credentials are no longer valid, and need
	// to be retrieved.
	IsExpired() bool
}

type Credentials struct {
	credentials Value
	provider    Provider

	sync.RWMutex
}

func NewCredentials(provider Provider) *Credentials {
	creds := &Credentials{
		provider: provider,
	}

	return creds
}

func (c *Credentials) Expire() {
	c.Lock()
	defer c.Unlock()

	c.credentials = Value{}
}

func (c *Credentials) Get() (Value, error) {
	if c.IsExpired() {
		if err := c.retrieve(); err != nil {
			return Value{}, err
		}
	}
	c.RLock()
	defer c.RUnlock()

	if !c.credentials.IsSet() {
		return Value{}, ErrMissingIncomplete
	}

	return c.credentials, nil
}

func (c *Credentials) IsExpired() bool {
	c.RLock()
	defer c.RUnlock()

	return (!c.credentials.IsSet() || c.provider.IsExpired())
}

func (c *Credentials) retrieve() error {
	c.Lock()
	defer c.Unlock()

	v, err := c.provider.Retrieve()
	if err != nil {
		return err
	}

	c.credentials = v

	return nil
}

// Reload asks the provider to re-read the credentials, if it is a Reloader (or a ChainProvider
// whose current provider is one). It does nothing for the other providers.
// New credentials are returned by the next Get.
func (c *Credentials) Reload(ctx context.Context) error {
	c.RLock()
	p := c.provider
	if chain, ok := p.(*ChainProvider); ok {
		p = chain.current
	}
	c.RUnlock()

	// The lock isn't held while reloading, so that Get doesn't wait for the source to be read.
	if r, ok := p.(Reloader); ok {
		return r.Reload(ctx)
	}

	return nil
}
