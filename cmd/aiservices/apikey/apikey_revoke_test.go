package apikey

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

type aiAPIKeyRevokeServer struct {
	server *httptest.Server
	keys   []v3.ListAIAPIKeysResponseEntry

	mu       sync.Mutex
	revoked  []v3.UUID
	revokeFN func(id v3.UUID) error
}

func newAIAPIKeyRevokeServer(t *testing.T) *aiAPIKeyRevokeServer {
	t.Helper()
	ts := &aiAPIKeyRevokeServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/api-key", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListAIAPIKeysResponse{AIAPIKeys: ts.keys})
	})
	mux.HandleFunc("/ai/api-key/", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if !strings.HasSuffix(p, "/revoke") || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := v3.UUID(strings.TrimSuffix(strings.TrimPrefix(p, "/ai/api-key/"), "/revoke"))
		ts.mu.Lock()
		ts.revoked = append(ts.revoked, id)
		fn := ts.revokeFN
		ts.mu.Unlock()
		if fn != nil {
			if err := fn(id); err != nil {
				testutils.WriteJSON(t, w, http.StatusNotFound, map[string]string{"message": err.Error()})
				return
			}
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-revoke-" + id), State: v3.OperationStateSuccess})
	})
	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-revoke"), State: v3.OperationStateSuccess})
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func (ts *aiAPIKeyRevokeServer) revokedIDs() []v3.UUID {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	out := make([]v3.UUID, len(ts.revoked))
	copy(out, ts.revoked)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func TestAIAPIKeyRevoke(t *testing.T) {
	now := time.Now()
	ts := newAIAPIKeyRevokeServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	ts.keys = []v3.ListAIAPIKeysResponseEntry{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", CreatedAT: now, UpdatedAT: now},
		{ID: v3.UUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), Name: "beta", CreatedAT: now, UpdatedAT: now},
	}

	// resolve a name and an ID, force skips the confirmation prompt
	c := &AIAPIKeyRevokeCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Keys:               []string{"alpha", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"},
		Force:              true,
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("api-key revoke: %v", err)
	}
	ids := ts.revokedIDs()
	if len(ids) != 2 || ids[0] != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" || ids[1] != "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" {
		t.Fatalf("expected both keys revoked, got %v", ids)
	}
}

func TestAIAPIKeyRevokeUnknownKey(t *testing.T) {
	now := time.Now()
	ts := newAIAPIKeyRevokeServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	ts.keys = []v3.ListAIAPIKeysResponseEntry{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", CreatedAT: now, UpdatedAT: now},
	}

	// without force an unknown key aborts the command
	c := &AIAPIKeyRevokeCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Keys:               []string{"missing"},
	}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatal("expected error for unknown key without force")
	}
	if got := ts.revokedIDs(); len(got) != 0 {
		t.Fatalf("expected no revocation, got %v", got)
	}

	// with force the unknown key is skipped with a warning
	c = &AIAPIKeyRevokeCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Keys:               []string{"missing", "alpha"},
		Force:              true,
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("api-key revoke with force: %v", err)
	}
	ids := ts.revokedIDs()
	if len(ids) != 1 || ids[0] != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("expected only alpha revoked, got %v", ids)
	}
}

func TestAIAPIKeyRevokeWithZone(t *testing.T) {
	now := time.Now()
	ts := newAIAPIKeyRevokeServer(t)
	defer ts.server.Close()

	controlMux := http.NewServeMux()
	controlMux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListZonesResponse{
			Zones: []v3.Zone{{Name: v3.ZoneName("z1"), APIEndpoint: v3.Endpoint(ts.server.URL)}},
		})
	})
	controlSrv := httptest.NewServer(controlMux)
	defer controlSrv.Close()
	testutils.SetupV3Client(t, controlSrv.URL)

	ts.keys = []v3.ListAIAPIKeysResponseEntry{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", CreatedAT: now, UpdatedAT: now},
	}

	c := &AIAPIKeyRevokeCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Keys:               []string{"alpha"},
		Force:              true,
		Zone:               "z1",
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("api-key revoke with zone: %v", err)
	}
	ids := ts.revokedIDs()
	if len(ids) != 1 || ids[0] != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("expected alpha revoked on the zoned endpoint, got %v", ids)
	}
}

func TestAIAPIKeyRevokeFailure(t *testing.T) {
	ts := newAIAPIKeyRevokeServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	now := time.Now()
	ts.keys = []v3.ListAIAPIKeysResponseEntry{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", CreatedAT: now, UpdatedAT: now},
	}

	ts.revokeFN = func(v3.UUID) error {
		return errors.New("cannot revoke")
	}
	c := &AIAPIKeyRevokeCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Keys:               []string{"alpha"},
		Force:              true,
	}
	err := c.CmdRun(nil, nil)
	if err == nil {
		t.Fatal("expected revocation failure")
	}
	if !errors.Is(err, v3.ErrNotFound) {
		t.Errorf("expected ErrNotFound sentinel, got %v", err)
	}
}
