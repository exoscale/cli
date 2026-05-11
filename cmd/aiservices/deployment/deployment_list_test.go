package deployment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

// zoneBehavior describes how a fake per-zone backend should respond.
type zoneBehavior int

const (
	zoneHealthy zoneBehavior = iota
	zoneTimeout
	zoneServerError
	zoneEmpty
)

// zoneFixture is one fake per-zone backend.
type zoneFixture struct {
	Name        string
	Behavior    zoneBehavior
	Deployments []v3.ListDeploymentsResponseEntry
	server      *httptest.Server
	calls       atomic.Int32
}

// multiZoneHarness wires together a fake control-plane server (which
// answers /zone) and one backend server per zone fixture (which answers
// /ai/deployment).
type multiZoneHarness struct {
	control *httptest.Server
	zones   []*zoneFixture
}

func (h *multiZoneHarness) close() {
	h.control.Close()
	for _, z := range h.zones {
		if z.server != nil {
			z.server.Close()
		}
	}
}

func newMultiZoneHarness(t *testing.T, fixtures []*zoneFixture, slowDelay time.Duration) *multiZoneHarness {
	t.Helper()
	h := &multiZoneHarness{zones: fixtures}

	// Build per-zone backends first so we know their URLs.
	for _, z := range fixtures {
		zone := z // capture
		mux := http.NewServeMux()
		mux.HandleFunc("/ai/deployment", func(w http.ResponseWriter, r *http.Request) {
			zone.calls.Add(1)
			switch zone.Behavior {
			case zoneTimeout:
				select {
				case <-r.Context().Done():
				case <-time.After(slowDelay):
				}
				return
			case zoneServerError:
				http.Error(w, "boom", http.StatusInternalServerError)
				return
			case zoneEmpty:
				writeJSON(t, w, http.StatusOK, v3.ListDeploymentsResponse{})
				return
			default:
				writeJSON(t, w, http.StatusOK, v3.ListDeploymentsResponse{Deployments: zone.Deployments})
			}
		})
		zone.server = httptest.NewServer(mux)
	}

	// Control plane: answer /zone with all fixture endpoints.
	controlMux := http.NewServeMux()
	controlMux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		zones := make([]v3.Zone, 0, len(fixtures))
		for _, z := range fixtures {
			zones = append(zones, v3.Zone{
				Name:        v3.ZoneName(z.Name),
				APIEndpoint: v3.Endpoint(z.server.URL),
			})
		}
		writeJSON(t, w, http.StatusOK, v3.ListZonesResponse{Zones: zones})
	})
	h.control = httptest.NewServer(controlMux)
	return h
}

func writeJSON(t *testing.T, w http.ResponseWriter, code int, v interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

func setupClient(t *testing.T, controlURL string) {
	t.Helper()
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(controlURL))
}

func sampleDeployments(zoneSuffix string, n int) []v3.ListDeploymentsResponseEntry {
	now := time.Now()
	out := make([]v3.ListDeploymentsResponseEntry, n)
	for i := 0; i < n; i++ {
		out[i] = v3.ListDeploymentsResponseEntry{
			ID:        v3.UUID(fmt.Sprintf("00000000-0000-0000-0000-%012d", i)),
			Name:      fmt.Sprintf("dep-%s-%d", zoneSuffix, i),
			State:     v3.ListDeploymentsResponseEntryStateReady,
			GpuType:   "gpua5000",
			GpuCount:  1,
			Replicas:  1,
			CreatedAT: now,
			UpdatedAT: now,
		}
	}
	return out
}

func runList(t *testing.T, zoneFilter v3.ZoneName) (stdout, stderr string, err error) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := &DeploymentListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Zone:               zoneFilter,
	}
	err = runDeploymentList(cmd, &outBuf, &errBuf)
	return outBuf.String(), errBuf.String(), err
}

func TestDeploymentList_AllHealthy(t *testing.T) {
	defer withTimeout(t, 5*time.Second)()
	zones := []*zoneFixture{
		{Name: "z1", Behavior: zoneHealthy, Deployments: sampleDeployments("z1", 2)},
		{Name: "z2", Behavior: zoneHealthy, Deployments: sampleDeployments("z2", 1)},
		{Name: "z3", Behavior: zoneEmpty},
	}
	h := newMultiZoneHarness(t, zones, 0)
	defer h.close()
	setupClient(t, h.control.URL)
	defer withFormat(t, "json")()

	stdout, stderr, err := runList(t, "")
	if err != nil {
		t.Fatalf("expected nil err, got %v (stderr=%s)", err, stderr)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}

	var rows []DeploymentListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, stdout)
	}
	if len(rows) != 3 {
		t.Errorf("want 3 rows, got %d", len(rows))
	}
	zoneCounts := map[v3.ZoneName]int{}
	for _, r := range rows {
		zoneCounts[r.Zone]++
	}
	if zoneCounts["z1"] != 2 || zoneCounts["z2"] != 1 || zoneCounts["z3"] != 0 {
		t.Errorf("unexpected zone distribution: %+v", zoneCounts)
	}
}

func TestDeploymentList_OneTimeout(t *testing.T) {
	defer withTimeout(t, 200*time.Millisecond)()
	zones := []*zoneFixture{
		{Name: "z1", Behavior: zoneHealthy, Deployments: sampleDeployments("z1", 1)},
		{Name: "z2", Behavior: zoneTimeout},
		{Name: "z3", Behavior: zoneHealthy, Deployments: sampleDeployments("z3", 1)},
	}
	h := newMultiZoneHarness(t, zones, 2*time.Second)
	defer h.close()
	setupClient(t, h.control.URL)
	defer withFormat(t, "json")()

	start := time.Now()
	stdout, stderr, err := runList(t, "")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected non-nil err on partial failure")
	}
	if !strings.Contains(err.Error(), "1 zone") {
		t.Errorf("err should mention 1 zone, got %q", err.Error())
	}
	if !strings.Contains(stderr, "warning: zone z2") {
		t.Errorf("stderr should warn about z2, got: %s", stderr)
	}
	if elapsed > 1*time.Second {
		t.Errorf("must not wait for the slow zone (elapsed=%s)", elapsed)
	}
	var rows []DeploymentListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, stdout)
	}
	if len(rows) != 2 {
		t.Errorf("want 2 healthy rows, got %d", len(rows))
	}
}

func TestDeploymentList_OneServerError(t *testing.T) {
	defer withTimeout(t, 5*time.Second)()
	zones := []*zoneFixture{
		{Name: "z1", Behavior: zoneHealthy, Deployments: sampleDeployments("z1", 1)},
		{Name: "z2", Behavior: zoneServerError},
	}
	h := newMultiZoneHarness(t, zones, 0)
	defer h.close()
	setupClient(t, h.control.URL)
	defer withFormat(t, "json")()

	stdout, stderr, err := runList(t, "")
	if err == nil {
		t.Fatal("expected non-nil err")
	}
	if !strings.Contains(stderr, "warning: zone z2") {
		t.Errorf("stderr should warn about z2, got: %s", stderr)
	}
	var rows []DeploymentListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, stdout)
	}
	if len(rows) != 1 {
		t.Errorf("want 1 healthy row, got %d", len(rows))
	}
}

func TestDeploymentList_AllFailed(t *testing.T) {
	defer withTimeout(t, 5*time.Second)()
	zones := []*zoneFixture{
		{Name: "z1", Behavior: zoneServerError},
		{Name: "z2", Behavior: zoneServerError},
	}
	h := newMultiZoneHarness(t, zones, 0)
	defer h.close()
	setupClient(t, h.control.URL)
	defer withFormat(t, "json")()

	stdout, stderr, err := runList(t, "")
	if err == nil {
		t.Fatal("expected non-nil err")
	}
	if !strings.Contains(stderr, "warning: zone z1") || !strings.Contains(stderr, "warning: zone z2") {
		t.Errorf("stderr should warn about both zones, got: %s", stderr)
	}
	got := strings.TrimSpace(stdout)
	if got != "[]" && got != "null" {
		t.Errorf("want empty json, got %q", got)
	}
}

func TestDeploymentList_ZoneFilter(t *testing.T) {
	defer withTimeout(t, 5*time.Second)()
	zones := []*zoneFixture{
		{Name: "z1", Behavior: zoneHealthy, Deployments: sampleDeployments("z1", 1)},
		{Name: "z2", Behavior: zoneHealthy, Deployments: sampleDeployments("z2", 1)},
	}
	h := newMultiZoneHarness(t, zones, 0)
	defer h.close()
	setupClient(t, h.control.URL)
	defer withFormat(t, "json")()

	stdout, _, err := runList(t, "z1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	var rows []DeploymentListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(rows) != 1 || rows[0].Zone != "z1" {
		t.Errorf("expected only z1 row, got %+v", rows)
	}
	if zones[1].calls.Load() != 0 {
		t.Errorf("z2 should not have been queried, calls=%d", zones[1].calls.Load())
	}
}

// withTimeout sets globalstate.RequestTimeout for the duration of a test.
func withTimeout(t *testing.T, d time.Duration) func() {
	t.Helper()
	prev := globalstate.RequestTimeout
	globalstate.RequestTimeout = d
	return func() { globalstate.RequestTimeout = prev }
}

// withFormat sets globalstate.OutputFormat for the duration of a test.
func withFormat(t *testing.T, f string) func() {
	t.Helper()
	prev := globalstate.OutputFormat
	globalstate.OutputFormat = f
	return func() { globalstate.OutputFormat = prev }
}
