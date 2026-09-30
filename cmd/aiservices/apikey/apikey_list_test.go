package apikey

import (
	"bytes"
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
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

// zoneBehavior describes how a fake per-zone backend should respond.
type apikeyZoneBehavior int

const (
	apikeyZoneHealthy apikeyZoneBehavior = iota
	apikeyZoneServerError
	apikeyZoneEmpty
)

// apikeyZoneFixture is one fake per-zone backend.
type apikeyZoneFixture struct {
	Name     string
	Behavior apikeyZoneBehavior
	Keys     []v3.ListAIAPIKeysResponseEntry
	server   *httptest.Server
	calls    atomic.Int32
}

// apikeyMultiZoneHarness wires together a fake control-plane server (which
// answers /zone) and one backend server per zone fixture (which answers
// /ai/api-key).
type apikeyMultiZoneHarness struct {
	control *httptest.Server
	zones   []*apikeyZoneFixture
}

func (h *apikeyMultiZoneHarness) close() {
	h.control.Close()
	for _, z := range h.zones {
		if z.server != nil {
			z.server.Close()
		}
	}
}

func newAPIKeyMultiZoneHarness(t *testing.T, fixtures []*apikeyZoneFixture) *apikeyMultiZoneHarness {
	t.Helper()
	h := &apikeyMultiZoneHarness{zones: fixtures}

	for _, z := range fixtures {
		zone := z
		mux := http.NewServeMux()
		mux.HandleFunc("/ai/api-key", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			zone.calls.Add(1)
			switch zone.Behavior {
			case apikeyZoneServerError:
				http.Error(w, "boom", http.StatusInternalServerError)
			case apikeyZoneEmpty:
				testutils.WriteJSON(t, w, http.StatusOK, v3.ListAIAPIKeysResponse{})
			default:
				testutils.WriteJSON(t, w, http.StatusOK, v3.ListAIAPIKeysResponse{AIAPIKeys: zone.Keys})
			}
		})
		z.server = httptest.NewServer(mux)
	}

	controlMux := http.NewServeMux()
	controlMux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		zones := make([]v3.Zone, 0, len(fixtures))
		for _, z := range fixtures {
			zones = append(zones, v3.Zone{
				Name:        v3.ZoneName(z.Name),
				APIEndpoint: v3.Endpoint(z.server.URL),
			})
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListZonesResponse{Zones: zones})
	})
	h.control = httptest.NewServer(controlMux)
	return h
}

func sampleAPIKeys(zoneSuffix string, n int) []v3.ListAIAPIKeysResponseEntry {
	now := time.Now()
	out := make([]v3.ListAIAPIKeysResponseEntry, n)
	for i := range out {
		revokedAt := now.Add(time.Duration(i) * time.Hour)
		out[i] = v3.ListAIAPIKeysResponseEntry{
			ID:             v3.UUID(fmt.Sprintf("00000000-0000-0000-0000-%012d", i)),
			Name:           fmt.Sprintf("key-%s-%d", zoneSuffix, i),
			Models:         &v3.AIAPIKeyModels{"all"},
			Deployments:    &v3.AIAPIKeyDeployments{},
			AllModels:      boolPtr(true),
			AllDeployments: boolPtr(false),
			CreatedAT:      now,
			UpdatedAT:      now,
			RevokedAT:      &revokedAt,
		}
	}
	return out
}

func runAPIKeyList(t *testing.T, zoneFilter v3.ZoneName) (stdout, stderr string, err error) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := &AIAPIKeyListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Zone:               zoneFilter,
	}
	err = runAIAPIKeyList(cmd, &outBuf, &errBuf)
	return outBuf.String(), errBuf.String(), err
}

func TestAIAPIKeyList_AllZones(t *testing.T) {
	defer withAPIKeyListTimeout(t, 5*time.Second)()
	h := newAPIKeyMultiZoneHarness(t, []*apikeyZoneFixture{
		{Name: "z1", Behavior: apikeyZoneHealthy},
		{Name: "z2", Behavior: apikeyZoneHealthy},
		{Name: "z3", Behavior: apikeyZoneEmpty},
	})
	h.zones[0].Keys = sampleAPIKeys("z1", 2)
	h.zones[1].Keys = sampleAPIKeys("z2", 1)
	defer h.close()
	testutils.SetupV3Client(t, h.control.URL)
	defer withAPIKeyListFormat(t, "json")()

	stdout, stderr, err := runAPIKeyList(t, "")
	if err != nil {
		t.Fatalf("expected nil err, got %v (stderr=%s)", err, stderr)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}

	var rows []AIAPIKeyListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, stdout)
	}
	if len(rows) != 3 {
		t.Fatalf("want 3 rows, got %d", len(rows))
	}

	zoneCounts := map[v3.ZoneName]int{}
	for _, r := range rows {
		zoneCounts[r.Zone]++
		if !strings.HasPrefix(r.Name, "key-"+string(r.Zone)) {
			t.Errorf("row %q not tagged with its zone %q", r.Name, r.Zone)
		}
	}
	if zoneCounts["z1"] != 2 || zoneCounts["z2"] != 1 || zoneCounts["z3"] != 0 {
		t.Errorf("unexpected zone distribution: %+v", zoneCounts)
	}
}

func TestAIAPIKeyList_ZoneServerError(t *testing.T) {
	defer withAPIKeyListTimeout(t, 5*time.Second)()
	h := newAPIKeyMultiZoneHarness(t, []*apikeyZoneFixture{
		{Name: "z1", Behavior: apikeyZoneHealthy},
		{Name: "z2", Behavior: apikeyZoneServerError},
	})
	h.zones[0].Keys = sampleAPIKeys("z1", 1)
	defer h.close()
	testutils.SetupV3Client(t, h.control.URL)
	defer withAPIKeyListFormat(t, "json")()

	stdout, stderr, err := runAPIKeyList(t, "")
	if err == nil {
		t.Fatal("expected non-nil err on partial failure")
	}
	if !strings.Contains(err.Error(), "1 zone") {
		t.Errorf("err should mention 1 zone, got %q", err.Error())
	}
	if !strings.Contains(stderr, "warning: zone z2") {
		t.Errorf("stderr should warn about z2, got: %s", stderr)
	}

	var rows []AIAPIKeyListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, stdout)
	}
	if len(rows) != 1 {
		t.Errorf("want 1 healthy row, got %d", len(rows))
	}
}

func TestAIAPIKeyList_ZoneFilter(t *testing.T) {
	defer withAPIKeyListTimeout(t, 5*time.Second)()
	h := newAPIKeyMultiZoneHarness(t, []*apikeyZoneFixture{
		{Name: "z1", Behavior: apikeyZoneHealthy},
		{Name: "z2", Behavior: apikeyZoneHealthy},
	})
	h.zones[0].Keys = sampleAPIKeys("z1", 1)
	h.zones[1].Keys = sampleAPIKeys("z2", 1)
	defer h.close()
	testutils.SetupV3Client(t, h.control.URL)
	defer withAPIKeyListFormat(t, "json")()

	stdout, _, err := runAPIKeyList(t, "z1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	var rows []AIAPIKeyListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, stdout)
	}
	if len(rows) != 1 || rows[0].Zone != "z1" {
		t.Errorf("expected only z1 row, got %+v", rows)
	}
	if h.zones[1].calls.Load() != 0 {
		t.Errorf("z2 should not have been queried, calls=%d", h.zones[1].calls.Load())
	}
}

func TestAIAPIKeyList_SortedByActiveAndName(t *testing.T) {
	defer withAPIKeyListTimeout(t, 5*time.Second)()
	now := time.Now()
	revokedAt := now
	h := newAPIKeyMultiZoneHarness(t, []*apikeyZoneFixture{
		{Name: "z1", Behavior: apikeyZoneHealthy, Keys: []v3.ListAIAPIKeysResponseEntry{
			{ID: v3.UUID("00000000-0000-0000-0000-000000000001"), Name: "bravo", CreatedAT: now, UpdatedAT: now},
			{ID: v3.UUID("00000000-0000-0000-0000-000000000002"), Name: "alpha", CreatedAT: now, UpdatedAT: now, RevokedAT: &revokedAt},
		}},
		{Name: "z2", Behavior: apikeyZoneHealthy, Keys: []v3.ListAIAPIKeysResponseEntry{
			{ID: v3.UUID("00000000-0000-0000-0000-000000000003"), Name: "zulu", CreatedAT: now, UpdatedAT: now},
			{ID: v3.UUID("00000000-0000-0000-0000-000000000004"), Name: "charlie", CreatedAT: now, UpdatedAT: now},
		}},
	})
	defer h.close()
	testutils.SetupV3Client(t, h.control.URL)
	defer withAPIKeyListFormat(t, "json")()

	stdout, _, err := runAPIKeyList(t, "")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	var rows []AIAPIKeyListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, stdout)
	}

	want := []string{"bravo", "charlie", "zulu", "alpha"}
	if len(rows) != len(want) {
		t.Fatalf("want %d rows, got %d", len(want), len(rows))
	}
	for i, name := range want {
		if rows[i].Name != name {
			t.Errorf("row %d: want %q, got %q", i, name, rows[i].Name)
		}
	}
	if rows[3].RevokedAt == "" {
		t.Errorf("expected last row to be the revoked key, got %+v", rows[3])
	}
}

func withAPIKeyListTimeout(t *testing.T, d time.Duration) func() {
	t.Helper()
	prev := globalstate.RequestTimeout
	globalstate.RequestTimeout = d
	return func() { globalstate.RequestTimeout = prev }
}

func withAPIKeyListFormat(t *testing.T, f string) func() {
	t.Helper()
	prev := globalstate.OutputFormat
	globalstate.OutputFormat = f
	return func() { globalstate.OutputFormat = prev }
}
