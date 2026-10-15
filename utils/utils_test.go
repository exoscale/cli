package utils

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
)

func TestParseInstanceType(t *testing.T) {
	testCases := []struct {
		instanceType   string
		expectedFamily v3.InstanceTypeFamily
		expectedSize   v3.InstanceTypeSize
	}{
		{"standard.large", v3.InstanceTypeFamily("standard"), v3.InstanceTypeSize("large")},
		{"gpu2.mega", v3.InstanceTypeFamily("gpu2"), v3.InstanceTypeSize("mega")},
		{"colossus", v3.InstanceTypeFamily("standard"), v3.InstanceTypeSize("colossus")},
		{"", v3.InstanceTypeFamily("standard"), v3.InstanceTypeSize("")},
		{"invalid-format", v3.InstanceTypeFamily("standard"), v3.InstanceTypeSize("invalid-format")},
	}

	for _, tc := range testCases {
		t.Run(tc.instanceType, func(t *testing.T) {
			result := ParseInstanceType(tc.instanceType)
			assert.Equal(t, tc.expectedFamily, result.Family)
			assert.Equal(t, tc.expectedSize, result.Size)
		})
	}
}

func TestWarningSink_ConcurrentAdd(t *testing.T) {
	var buf bytes.Buffer
	s := NewWarningSinkTo(&buf)

	const N = 50
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			s.Add("zone-%d failed", i)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, N, s.Len())
	s.Flush()
	assert.Equal(t, 0, s.Len(), "Flush should drain")

	out := buf.String()
	for i := 0; i < N; i++ {
		assert.Contains(t, out, fmt.Sprintf("warning: zone-%d failed", i))
	}
}

func TestSpinner_StartStopOnNonTTYIsNoop(t *testing.T) {
	var buf bytes.Buffer
	// Real Spinner with a non-TTY writer: Start should be a silent
	// no-op so scripted runs don't get spinner spam.
	s := NewSpinner()
	s.SetWriter(&buf)
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	if buf.Len() != 0 {
		t.Errorf("non-TTY spinner should produce no output, got %q", buf.String())
	}
}

func TestWarningSink_FlushIsIdempotent(t *testing.T) {
	var buf bytes.Buffer
	s := NewWarningSinkTo(&buf)
	s.Add("once")
	s.Flush()
	s.Flush()
	assert.Equal(t, "warning: once\n", buf.String())
}

func TestForEveryZoneAsync_FastZoneNotBlockedBySlow(t *testing.T) {
	// Server A returns immediately. Server B blocks longer than the
	// per-zone timeout. We assert that A's work landed and B counted
	// as failed without holding up the whole call.
	const timeout = 200 * time.Millisecond
	const slowDelay = 800 * time.Millisecond

	fastSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"deployments":[]}`))
	}))
	defer fastSrv.Close()

	slowSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(slowDelay):
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"deployments":[]}`))
	}))
	defer slowSrv.Close()

	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	assert.NoError(t, err)

	zones := []v3.Zone{
		{Name: v3.ZoneName("fast"), APIEndpoint: v3.Endpoint(fastSrv.URL)},
		{Name: v3.ZoneName("slow"), APIEndpoint: v3.Endpoint(slowSrv.URL)},
	}

	var fastDone atomic.Bool
	var buf bytes.Buffer
	sink := NewWarningSinkTo(&buf)

	start := time.Now()
	failed := ForEveryZoneAsync(context.Background(), zones, timeout, sink, false,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			_, err := zc.ListDeployments(ctx)
			if err != nil {
				return err
			}
			if zone.Name == "fast" {
				fastDone.Store(true)
			}
			return nil
		})
	elapsed := time.Since(start)

	assert.True(t, fastDone.Load(), "fast zone should have completed")
	assert.Equal(t, 1, failed, "exactly one zone should have failed (slow)")
	assert.Equal(t, 1, sink.Len(), "sink should hold one warning")

	// Should have taken roughly the timeout, not slowDelay.
	assert.Less(t, elapsed, slowDelay, "must not wait for the slow zone")

	sink.Flush()
	assert.Contains(t, buf.String(), "zone slow")
}

func TestZoneLookupFallback(t *testing.T) {
	// countingServer answers the zone list with status, and counts the requests it gets.
	countingServer := func(status int, hits *atomic.Int32) *httptest.Server {
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

	var defaultHits, fallbackHits atomic.Int32
	defaultSrv := countingServer(http.StatusBadGateway, &defaultHits)
	fallbackSrv := countingServer(http.StatusOK, &fallbackHits)

	oldAccount, oldFallback, oldQuiet := account.CurrentAccount, fallbackEndpoint, globalstate.Quiet
	t.Cleanup(func() {
		account.CurrentAccount, fallbackEndpoint, globalstate.Quiet = oldAccount, oldFallback, oldQuiet
		lookupEndpoint.Store(nil)
	})
	fallbackEndpoint = func() v3.Endpoint { return v3.Endpoint(fallbackSrv.URL + "/v2") }
	globalstate.Quiet = true

	// No egoscale retries, so that the default endpoint fails right away.
	newClient := func(endpoint string) *v3.Client {
		client, err := v3.NewClient(
			credentials.NewStaticCredentials("key", "secret"),
			v3.ClientOptWithHTTPClient(&http.Client{}),
			v3.ClientOptWithEndpoint(v3.Endpoint(endpoint+"/v2")),
		)
		require.NoError(t, err)
		return client
	}
	client := newClient(defaultSrv.URL)

	reset := func(endpoint string) {
		defaultHits.Store(0)
		fallbackHits.Store(0)
		lookupEndpoint.Store(nil)
		account.CurrentAccount = &account.Account{Endpoint: endpoint}
	}

	t.Run("default endpoint down, next lookups go to the fallback", func(t *testing.T) {
		reset("")

		endpoint, err := GetZoneAPIEndpointV3(context.Background(), client, "ch-dk-2")
		require.NoError(t, err)
		assert.Equal(t, v3.CHDk2, endpoint)

		_, err = ListZonesV3(context.Background(), client)
		require.NoError(t, err)
		assert.EqualValues(t, 1, defaultHits.Load())
		assert.EqualValues(t, 2, fallbackHits.Load())
	})

	t.Run("unknown zone", func(t *testing.T) {
		reset("")

		_, err := GetZoneAPIEndpointV3(context.Background(), newClient(fallbackSrv.URL), "xx-nowhere-1")
		require.ErrorIs(t, err, v3.ErrNotFound)
		// The public then the signed zone list, on the given endpoint only.
		assert.EqualValues(t, 2, fallbackHits.Load())
	})

	t.Run("configured endpoint down", func(t *testing.T) {
		reset(defaultSrv.URL + "/v2")

		_, err := GetZoneAPIEndpointV3(context.Background(), client, "ch-dk-2")
		require.Error(t, err)

		_, err = ListZonesV3(context.Background(), client)
		require.Error(t, err)
		assert.EqualValues(t, 0, fallbackHits.Load())
	})
}

func TestFallbackEndpoint(t *testing.T) {
	for range 100 {
		assert.NotEqual(t, v3.CHGva2, fallbackEndpoint())
	}
}
