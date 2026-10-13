package zonelookup

import (
	"context"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

// zoneServer serves the zone list with the given status, and counts the requests it gets.
func zoneServer(t *testing.T, status int, hits *atomic.Int32) v3.Endpoint {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"zones":[{"name":"ch-dk-2","api-endpoint":"https://api-ch-dk-2.exoscale.com/v2"}]}`))
	}))
	t.Cleanup(srv.Close)

	return v3.Endpoint(srv.URL + "/v2")
}

// downEndpoint returns the endpoint of a server that no longer listens.
func downEndpoint(t *testing.T) v3.Endpoint {
	t.Helper()
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	return v3.Endpoint(srv.URL + "/v2")
}

// setup returns a client on endpoint, falling back to fallbacks, with endpoint
// configured in the account or not.
func setup(t *testing.T, configured bool, endpoint v3.Endpoint, fallbacks ...v3.Endpoint) *v3.Client {
	t.Helper()

	oldAccount, oldFallbacks := account.CurrentAccount, fallbackEndpoints
	t.Cleanup(func() {
		account.CurrentAccount, fallbackEndpoints, available = oldAccount, oldFallbacks, ""
	})

	account.CurrentAccount = &account.Account{}
	if configured {
		account.CurrentAccount.Endpoint = string(endpoint)
	}
	fallbackEndpoints = fallbacks
	globalstate.Quiet = true

	// No egoscale retries, so that a down endpoint fails right away.
	client, err := v3.NewClient(
		credentials.NewStaticCredentials("key", "secret"),
		v3.ClientOptWithHTTPClient(&http.Client{}),
		v3.ClientOptWithEndpoint(endpoint),
	)
	require.NoError(t, err)

	return client
}

func TestLookup(t *testing.T) {
	ctx := context.Background()

	t.Run("default endpoint answers", func(t *testing.T) {
		var first, second atomic.Int32
		client := setup(t, false, zoneServer(t, http.StatusOK, &first), zoneServer(t, http.StatusOK, &second))

		_, err := ListZones(ctx, client)
		require.NoError(t, err)
		require.EqualValues(t, 1, first.Load())
		require.EqualValues(t, 0, second.Load())
	})

	t.Run("default endpoint unreachable", func(t *testing.T) {
		var hits atomic.Int32
		client := setup(t, false, downEndpoint(t), zoneServer(t, http.StatusOK, &hits))

		endpoint, err := GetZoneAPIEndpoint(ctx, client, "ch-dk-2")
		require.NoError(t, err)
		require.Equal(t, v3.CHDk2, endpoint)
		require.EqualValues(t, 1, hits.Load())
	})

	t.Run("default endpoint answers 5xx, next lookups go to the fallback", func(t *testing.T) {
		var first, second atomic.Int32
		client := setup(t, false, zoneServer(t, http.StatusServiceUnavailable, &first), zoneServer(t, http.StatusOK, &second))

		for range 2 {
			_, err := ListZones(ctx, client)
			require.NoError(t, err)
		}
		require.EqualValues(t, 1, first.Load())
		require.EqualValues(t, 2, second.Load())
	})

	t.Run("default endpoint answers 4xx", func(t *testing.T) {
		var first, second atomic.Int32
		client := setup(t, false, zoneServer(t, http.StatusForbidden, &first), zoneServer(t, http.StatusOK, &second))

		_, err := ListZones(ctx, client)
		require.ErrorIs(t, err, v3.ErrForbidden)
		require.EqualValues(t, 0, second.Load())
	})

	t.Run("unknown zone", func(t *testing.T) {
		var first, second atomic.Int32
		client := setup(t, false, zoneServer(t, http.StatusOK, &first), zoneServer(t, http.StatusOK, &second))

		_, err := GetZoneAPIEndpoint(ctx, client, "xx-nowhere-1")
		require.ErrorIs(t, err, v3.ErrNotFound)
		require.EqualValues(t, 0, second.Load())
	})

	t.Run("configured endpoint unreachable", func(t *testing.T) {
		var hits atomic.Int32
		client := setup(t, true, downEndpoint(t), zoneServer(t, http.StatusOK, &hits))

		_, err := ListZones(ctx, client)
		require.Error(t, err)
		require.EqualValues(t, 0, hits.Load())
	})

	t.Run("all endpoints unavailable", func(t *testing.T) {
		var hits atomic.Int32
		client := setup(t, false, downEndpoint(t), zoneServer(t, http.StatusBadGateway, &hits))

		_, err := ListZones(ctx, client)
		require.ErrorIs(t, err, v3.ErrBadGateway)
		require.EqualValues(t, 1, hits.Load())
	})
}

func TestFallbackEndpoints(t *testing.T) {
	require.NotContains(t, fallbackEndpoints, v3.CHGva2)
	require.ElementsMatch(t, slices.Collect(maps.Values(v3.ZoneEndpoints)), append(fallbackEndpoints, v3.CHGva2))
}
