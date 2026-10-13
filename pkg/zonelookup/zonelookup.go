// Package zonelookup looks zones up on the Exoscale API without depending on a
// single zone being available.
package zonelookup

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

// attemptTimeout bounds the lookup on each endpoint, so that an unreachable
// endpoint doesn't hold the CLI before the next one is tried.
const attemptTimeout = 5 * time.Second

// fallbackEndpoints are asked for the zones, in order, when the default endpoint
// (ch-gva-2) is unavailable. The zones themselves always come from the API.
var fallbackEndpoints = func() []v3.Endpoint {
	var endpoints []v3.Endpoint
	for _, z := range slices.Sorted(maps.Keys(v3.ZoneEndpoints)) {
		if e := v3.ZoneEndpoints[z]; e != v3.CHGva2 {
			endpoints = append(endpoints, e)
		}
	}
	return endpoints
}()

var (
	mu sync.Mutex
	// available is the fallback endpoint that answered last, tried first afterwards.
	available v3.Endpoint
)

// ListZones lists the zones, see lookup.
func ListZones(ctx context.Context, client *v3.Client) (*v3.ListZonesResponse, error) {
	return lookup(ctx, client, func(ctx context.Context, c *v3.Client) (*v3.ListZonesResponse, error) {
		return c.ListZones(ctx)
	})
}

// GetZoneAPIEndpoint returns the API endpoint of a zone, see lookup.
func GetZoneAPIEndpoint(ctx context.Context, client *v3.Client, zone v3.ZoneName) (v3.Endpoint, error) {
	return lookup(ctx, client, func(ctx context.Context, c *v3.Client) (v3.Endpoint, error) {
		return c.GetZoneAPIEndpoint(ctx, zone)
	})
}

// lookup runs fn on client and, when no API endpoint is configured and client's
// endpoint is unavailable, on the other zone endpoints in turn: every zone
// endpoint returns the same zones. A configured endpoint (profile or
// EXOSCALE_API_ENDPOINT) is an explicit choice and is used without fallback.
func lookup[T any](ctx context.Context, client *v3.Client, fn func(context.Context, *v3.Client) (T, error)) (T, error) {
	if account.CurrentAccount != nil && account.CurrentAccount.Endpoint != "" {
		return fn(ctx, client)
	}

	mu.Lock()
	last := available
	mu.Unlock()

	// "" stands for client's own endpoint, skipped once a fallback answered.
	endpoints := []v3.Endpoint{last}
	for _, e := range fallbackEndpoints {
		if e != last {
			endpoints = append(endpoints, e)
		}
	}

	var (
		res T
		err error
	)
	for _, e := range endpoints {
		c := client
		if e != "" {
			c = client.WithEndpoint(e)
		}

		attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
		res, err = fn(attemptCtx, c)
		cancel()

		if !unavailable(err) {
			if e != "" {
				remember(e)
			}
			return res, err
		}
		if ctx.Err() != nil {
			break
		}
	}

	return res, err
}

// unavailable reports whether err means the endpoint didn't answer, or failed
// to: any other answer would be the same on every endpoint.
func unavailable(err error) bool {
	var uerr *url.Error

	return errors.As(err, &uerr) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, v3.ErrInternalServerError) ||
		errors.Is(err, v3.ErrBadGateway) ||
		errors.Is(err, v3.ErrServiceUnavailable) ||
		errors.Is(err, v3.ErrGatewayTimeout)
}

func remember(endpoint v3.Endpoint) {
	mu.Lock()
	defer mu.Unlock()

	if available == "" && !globalstate.Quiet {
		fmt.Fprintf(os.Stderr, "warning: the default API endpoint is unavailable, looking zones up on %s\n", endpoint)
	}
	available = endpoint
}
