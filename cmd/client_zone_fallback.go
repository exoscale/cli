package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/hashicorp/go-retryablehttp"
)

// zoneListAttemptTimeout bounds each zone list attempt, so that an unreachable
// endpoint doesn't hold the CLI before the next one is tried.
const zoneListAttemptTimeout = 10 * time.Second

// zoneListEndpoints lists the public API endpoints the zone list is looked up on,
// in order, when no API endpoint is configured.
var zoneListEndpoints = []v3.Endpoint{
	v3.CHGva2,
	v3.CHDk2,
	v3.DEFra1,
	v3.DEMuc1,
	v3.ATVie1,
	v3.ATVie2,
	v3.BGSof1,
	v3.HrZag1,
}

// zoneListFallbackTransport sends the zone list request (GET /v2/zone) made on
// the first endpoint to the next one when it is unreachable or answers a 5xx,
// so that the CLI doesn't depend on a single zone to find the others.
// Every zonal endpoint returns the same zone list. Any other request goes
// through next unchanged, so a mutating request is never replayed.
type zoneListFallbackTransport struct {
	next      http.RoundTripper
	attempt   *http.Client
	endpoints []*url.URL

	mu      sync.Mutex
	current int // index in endpoints of the last endpoint that answered
}

// newZoneListFallbackHTTPClient returns the HTTP client egoscale uses by default
// (retrying on transient errors), with the zone list fallback on top.
func newZoneListFallbackHTTPClient() (*http.Client, error) {
	rc := retryablehttp.NewClient()
	rc.Logger = nil

	transport, err := newZoneListFallbackTransport(rc.StandardClient().Transport, zoneListEndpoints)
	if err != nil {
		return nil, err
	}

	return &http.Client{Transport: transport}, nil
}

func newZoneListFallbackTransport(next http.RoundTripper, endpoints []v3.Endpoint) (*zoneListFallbackTransport, error) {
	t := &zoneListFallbackTransport{
		next:    next,
		attempt: &http.Client{Transport: http.DefaultTransport, Timeout: zoneListAttemptTimeout},
	}

	for _, e := range endpoints {
		u, err := url.Parse(string(e))
		if err != nil {
			return nil, fmt.Errorf("parse endpoint %q: %w", e, err)
		}
		t.endpoints = append(t.endpoints, u)
	}

	return t, nil
}

func (t *zoneListFallbackTransport) isZoneList(req *http.Request) bool {
	first := t.endpoints[0]

	return req.Method == http.MethodGet &&
		req.URL.Host == first.Host &&
		strings.TrimSuffix(req.URL.Path, "/") == first.Path+"/zone"
}

func (t *zoneListFallbackTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.endpoints) == 0 || !t.isZoneList(req) {
		return t.next.RoundTrip(req)
	}

	t.mu.Lock()
	start := t.current
	t.mu.Unlock()

	var (
		resp *http.Response
		err  error
	)
	for i := range t.endpoints {
		idx := (start + i) % len(t.endpoints)

		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		resp, err = t.attempt.Do(withEndpoint(req, t.endpoints[idx]))
		if err == nil && resp.StatusCode < http.StatusInternalServerError {
			t.switchTo(idx)
			return resp, nil
		}

		if req.Context().Err() != nil {
			break
		}
	}

	// The caller's http.Client wraps the error into a *url.Error already.
	var uerr *url.Error
	if errors.As(err, &uerr) {
		err = uerr.Err
	}

	return resp, err
}

func (t *zoneListFallbackTransport) switchTo(idx int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if idx == t.current {
		return
	}

	if !globalstate.Quiet {
		fmt.Fprintf(os.Stderr, "warning: %s is unavailable, listing zones from %s\n",
			t.endpoints[t.current].Host, t.endpoints[idx].Host)
	}
	t.current = idx
}

func withEndpoint(req *http.Request, endpoint *url.URL) *http.Request {
	r := req.Clone(req.Context())
	r.URL.Scheme = endpoint.Scheme
	r.URL.Host = endpoint.Host
	r.Host = ""

	return r
}
