package model

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

type modelListTestServer struct {
	server        *httptest.Server
	models        []v3.ListModelsResponseEntry
	zoneListCount atomic.Int32

	mu              sync.Mutex
	visibilityCalls []string
}

func (ts *modelListTestServer) lastVisibility() (string, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if len(ts.visibilityCalls) == 0 {
		return "", false
	}
	return ts.visibilityCalls[len(ts.visibilityCalls)-1], true
}

func newModelListTestServer(t *testing.T) *modelListTestServer {
	ts := &modelListTestServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/model", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if v := r.URL.Query().Get("visibility"); v != "" {
				ts.mu.Lock()
				ts.visibilityCalls = append(ts.visibilityCalls, v)
				ts.mu.Unlock()
			}
			testutils.WriteJSON(t, w, http.StatusOK, v3.ListModelsResponse{Models: ts.models})
		case http.MethodPost:
			testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-model-create"), State: v3.OperationStateSuccess})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/ai/model/", func(w http.ResponseWriter, r *http.Request) {
		id := path.Base(r.URL.Path)
		if r.Method == http.MethodGet {
			for _, m := range ts.models {
				if string(m.ID) == id {
					testutils.WriteJSON(t, w, http.StatusOK, v3.GetModelResponse{ID: m.ID, Name: m.Name, State: v3.GetModelResponseState(m.State), ModelSize: m.ModelSize, CreatedAT: m.CreatedAT, UpdatedAT: m.UpdatedAT})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodDelete {
			testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-model-delete"), State: v3.OperationStateSuccess})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		ts.zoneListCount.Add(1)
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListZonesResponse{Zones: []v3.Zone{{APIEndpoint: v3.Endpoint(ts.server.URL), Name: v3.ZoneName("test-zone")}}})
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func runModelListTest(t *testing.T, zoneFilter v3.ZoneName, visibility v3.ListModelsResponseEntryVisibility) (stdout, stderr string, err error) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := &ModelListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Zone:               zoneFilter,
		Visibility:         visibility,
	}
	err = runModelList(cmd, &outBuf, &errBuf)
	return outBuf.String(), errBuf.String(), err
}

func TestModelList(t *testing.T) {
	ts := newModelListTestServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	now := time.Now()
	ts.models = []v3.ListModelsResponseEntry{
		{ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1", State: v3.ListModelsResponseEntryStateReady, ModelSize: 0, CreatedAT: now, UpdatedAT: now},
		{ID: v3.UUID("22222222-2222-2222-2222-222222222222"), Name: "m2", State: v3.ListModelsResponseEntryStateCreating, ModelSize: 1024 * 1024 * 1024, CreatedAT: now, UpdatedAT: now},
	}
	defer withFormat(t, "json")()

	stdout, _, err := runModelListTest(t, "", "")
	if err != nil {
		t.Fatalf("model list: %v", err)
	}

	var rows []ModelListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, stdout)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 models, got %d", len(rows))
	}
	for _, m := range rows {
		if m.Zone != "test-zone" {
			t.Errorf("expected zone %q, got %q", "test-zone", m.Zone)
		}
		if m.Name == "m1" && m.ModelSize != "" {
			t.Errorf("expected m1 size empty, got %q", m.ModelSize)
		}
		if m.Name == "m2" && m.ModelSize != "1.0 GiB" {
			t.Errorf("expected m2 size 1.0 GiB, got %q", m.ModelSize)
		}
	}
}

func TestModelListUsesZone(t *testing.T) {
	ts := newModelListTestServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	cmd := &ModelListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Zone: v3.ZoneName("test-zone")}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model list: %v", err)
	}
	if ts.zoneListCount.Load() == 0 {
		t.Fatalf("expected zone list endpoint to be called")
	}
}

func TestModelListCmd_CmdShort(t *testing.T) {
	cmd := &ModelListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	short := cmd.CmdShort()
	if short == "" {
		t.Fatal("CmdShort() returned empty string")
	}
}

func TestModelList_ZoneEmpty(t *testing.T) {
	ts := newModelListTestServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)
	defer withFormat(t, "json")()
	ts.models = nil

	stdout, _, err := runModelListTest(t, "", "")
	if err != nil {
		t.Fatalf("model list: %v", err)
	}

	var rows []ModelListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, stdout)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 models, got %d", len(rows))
	}
}

func TestModelListVisibilityFilter(t *testing.T) {
	ts := newModelListTestServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)
	defer withFormat(t, "json")()

	stdout, _, err := runModelListTest(t, "", v3.ListModelsResponseEntryVisibilityPublic)
	if err != nil {
		t.Fatalf("model list with visibility: %v", err)
	}

	var rows []ModelListItemOutput
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, stdout)
	}

	vis, ok := ts.lastVisibility()
	if !ok {
		t.Fatal("expected visibility query parameter to be sent")
	}
	if vis != "public" {
		t.Errorf("expected visibility %q, got %q", "public", vis)
	}
}

func TestModelListVisibilityPrivate(t *testing.T) {
	ts := newModelListTestServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)
	defer withFormat(t, "json")()

	if _, _, err := runModelListTest(t, "", v3.ListModelsResponseEntryVisibilityPrivate); err != nil {
		t.Fatalf("model list with private visibility: %v", err)
	}

	vis, ok := ts.lastVisibility()
	if !ok {
		t.Fatal("expected visibility query parameter to be sent")
	}
	if vis != "private" {
		t.Errorf("expected visibility %q, got %q", "private", vis)
	}
}

func TestModelListInvalidVisibility(t *testing.T) {
	ts := newModelListTestServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	_, _, err := runModelListTest(t, "", v3.ListModelsResponseEntryVisibility("bogus"))
	if err == nil {
		t.Fatal("expected error for invalid visibility")
	}
	if _, ok := ts.lastVisibility(); ok {
		t.Error("expected no model list call for invalid visibility")
	}
}

// withFormat sets globalstate.OutputFormat for the duration of a test.
func withFormat(t *testing.T, f string) func() {
	t.Helper()
	prev := globalstate.OutputFormat
	globalstate.OutputFormat = f
	return func() { globalstate.OutputFormat = prev }
}
