package exoapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

type countingTransport struct{ calls int }

func (c *countingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.calls++
	return http.DefaultTransport.RoundTrip(r)
}

func newTestClient(t *testing.T, opts ...v3.ClientOpt) Client {
	t.Helper()
	client, err := v3.NewClient(credentials.NewStaticCredentials("key", "secret"), opts...)
	assert.NoError(t, err)
	return New(client)
}

func TestAdapterWith(t *testing.T) {
	tests := []struct {
		name               string
		configure          func(c Client, transport http.RoundTripper) Client
		wantTransportCalls int
	}{
		{
			name:      "WithEndpoint targets the endpoint",
			configure: func(c Client, _ http.RoundTripper) Client { return c },
		},
		{
			name: "WithHttpClient uses the HTTP client",
			configure: func(c Client, transport http.RoundTripper) Client {
				return c.WithHttpClient(&http.Client{Transport: transport})
			},
			wantTransportCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				hits++
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"zones":[]}`))
			}))
			t.Cleanup(srv.Close)
			transport := &countingTransport{}

			client := tt.configure(newTestClient(t).WithEndpoint(v3.Endpoint(srv.URL)), transport)
			_, err := client.ListZones(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, 1, hits)
			assert.Equal(t, tt.wantTransportCalls, transport.calls)
		})
	}
}

func TestUserAgent(t *testing.T) {
	const userAgent = "exocli/test"

	tests := []struct {
		name   string
		client func(t *testing.T) Client
		want   string
	}{
		{
			name: "client created by New",
			client: func(t *testing.T) Client {
				return newTestClient(t, v3.ClientOptWithUserAgent(userAgent))
			},
			want: userAgent,
		},
		{
			name: "survives With calls",
			client: func(t *testing.T) Client {
				return newTestClient(t, v3.ClientOptWithUserAgent(userAgent)).WithEndpoint("https://api.invalid")
			},
			want: userAgent,
		},
		{
			name:   "nil client",
			client: func(*testing.T) Client { return nil },
		},
		{
			name:   "other implementation",
			client: func(*testing.T) Client { return struct{ Client }{} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserAgent(tt.client(t))
			if tt.want == "" {
				assert.Empty(t, got)
				return
			}
			assert.Contains(t, got, tt.want)
		})
	}
}
