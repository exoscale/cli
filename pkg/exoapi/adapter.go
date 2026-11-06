package exoapi

import (
	"net/http"
	"reflect"

	v3 "github.com/exoscale/egoscale/v3"
)

var _ Client = adapter{}

// adapter makes a *v3.Client satisfy Client: its With* methods return a
// *v3.Client, while the interface requires them to return a Client.
type adapter struct {
	*v3.Client
}

// New returns c as a Client.
func New(c *v3.Client) Client {
	return adapter{Client: c}
}

// WithEndpoint returns a copy of the client using the given endpoint.
func (a adapter) WithEndpoint(endpoint v3.Endpoint) Client {
	return adapter{Client: a.Client.WithEndpoint(endpoint)}
}

// WithHttpClient returns a copy of the client using the given HTTP client.
func (a adapter) WithHttpClient(client *http.Client) Client {
	return adapter{Client: a.Client.WithHttpClient(client)}
}

// UserAgent returns the User-Agent of a client created by New, and an empty
// string for any other implementation. egoscale doesn't expose it, hence the
// reflection.
func UserAgent(c Client) string {
	a, ok := c.(adapter)
	if !ok {
		return ""
	}

	return reflect.ValueOf(*a.Client).FieldByName("userAgent").String()
}
