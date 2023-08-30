package v3

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
)

// Version string will be embeded into User-Agent header.
// Real value should be set with compiler flag.
const version = "dev"

// "User-Agent" HTTP request header added to outgoing HTTP requests.
var defaultUserAgent = fmt.Sprintf("egoscale/%s (%s; %s/%s)",
	version,
	runtime.Version(),
	runtime.GOOS,
	runtime.GOARCH)

// UserAgentProvider is an request interceptor that adds "User-Agent" HTTP request header.
// It is used by default in the Client.
type UserAgentProvider struct {
	prefix string
}

// NewUserAgentProvider creates a UserAgentProvider request interceptor.
// User-Agent is always suffixed with library version and host info, for example:
//
//	"egoscale/0.100.2 (go1.20.4; linux/amd64)"
//
// Prefix provided may be empty, in which case .
func NewUserAgentProvider(prefix string) *UserAgentProvider {
	return &UserAgentProvider{
		prefix: prefix,
	}
}

// Intercept will add a "User-Agent" header to API call.
func (p *UserAgentProvider) Intercept(ctx context.Context, req *http.Request) error {
	ua := defaultUserAgent
	if p.prefix != "" {
		ua = p.prefix + " " + ua
	}
	req.Header.Add("User-Agent", ua)

	return nil
}
