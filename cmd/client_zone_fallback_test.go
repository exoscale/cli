package cmd

import (
	"context"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

// zoneServer serves the zone list with the given status, and counts the requests it gets.
func zoneServer(t *testing.T, status int, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"zones":[{"name":"ch-dk-2","api-endpoint":"https://api-ch-dk-2.exoscale.com/v2"}]}`))
	}))
	t.Cleanup(srv.Close)

	return srv
}

// downEndpoint returns the endpoint of a server that no longer listens.
func downEndpoint(t *testing.T) v3.Endpoint {
	t.Helper()
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	return v3.Endpoint(srv.URL + "/v2")
}

func fallbackClient(t *testing.T, endpoints ...v3.Endpoint) *v3.Client {
	t.Helper()
	globalstate.Quiet = true

	transport, err := newZoneListFallbackTransport(http.DefaultTransport, endpoints)
	require.NoError(t, err)

	client, err := v3.NewClient(
		credentials.NewStaticCredentials("key", "secret"),
		v3.ClientOptWithHTTPClient(&http.Client{Transport: transport}),
		v3.ClientOptWithEndpoint(endpoints[0]),
	)
	require.NoError(t, err)

	return client
}

func TestZoneListFallback(t *testing.T) {
	ctx := context.Background()

	t.Run("first endpoint answers", func(t *testing.T) {
		var first, second atomic.Int32
		client := fallbackClient(t,
			v3.Endpoint(zoneServer(t, http.StatusOK, &first).URL+"/v2"),
			v3.Endpoint(zoneServer(t, http.StatusOK, &second).URL+"/v2"),
		)

		_, err := client.ListZones(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 1, first.Load())
		require.EqualValues(t, 0, second.Load())
	})

	t.Run("first endpoint unreachable", func(t *testing.T) {
		var hits atomic.Int32
		client := fallbackClient(t, downEndpoint(t), v3.Endpoint(zoneServer(t, http.StatusOK, &hits).URL+"/v2"))

		endpoint, err := client.GetZoneAPIEndpoint(ctx, "ch-dk-2")
		require.NoError(t, err)
		require.Equal(t, v3.CHDk2, endpoint)
		require.EqualValues(t, 1, hits.Load())
	})

	t.Run("first endpoint answers 5xx, next lookups stick to the fallback", func(t *testing.T) {
		var first, second atomic.Int32
		client := fallbackClient(t,
			v3.Endpoint(zoneServer(t, http.StatusServiceUnavailable, &first).URL+"/v2"),
			v3.Endpoint(zoneServer(t, http.StatusOK, &second).URL+"/v2"),
		)

		for range 2 {
			_, err := client.ListZones(ctx)
			require.NoError(t, err)
		}
		require.EqualValues(t, 1, first.Load())
		require.EqualValues(t, 2, second.Load())
	})

	t.Run("first endpoint answers 4xx", func(t *testing.T) {
		var first, second atomic.Int32
		client := fallbackClient(t,
			v3.Endpoint(zoneServer(t, http.StatusForbidden, &first).URL+"/v2"),
			v3.Endpoint(zoneServer(t, http.StatusOK, &second).URL+"/v2"),
		)

		_, err := client.ListZones(ctx)
		require.Error(t, err)
		require.EqualValues(t, 1, first.Load())
		require.EqualValues(t, 0, second.Load())
	})

	t.Run("all endpoints unavailable", func(t *testing.T) {
		var hits atomic.Int32
		client := fallbackClient(t, downEndpoint(t), v3.Endpoint(zoneServer(t, http.StatusBadGateway, &hits).URL+"/v2"))

		_, err := client.ListZones(ctx)
		require.Error(t, err)
		require.EqualValues(t, 1, hits.Load())
	})

	t.Run("other requests are not replayed", func(t *testing.T) {
		var hits atomic.Int32
		client := fallbackClient(t, downEndpoint(t), v3.Endpoint(zoneServer(t, http.StatusOK, &hits).URL+"/v2"))

		_, err := client.ListInstances(ctx)
		require.Error(t, err)
		_, err = client.DeleteInstance(ctx, "6d2d2a2d-6a4e-4f0e-a5b6-0f6c3e7f2b1a")
		require.Error(t, err)
		require.EqualValues(t, 0, hits.Load())
	})
}

func TestZoneListEndpoints(t *testing.T) {
	endpoints := zoneListEndpoints()

	require.Equal(t, v3.CHGva2, endpoints[0])
	require.ElementsMatch(t, slices.Collect(maps.Values(v3.ZoneEndpoints)), endpoints)
}
