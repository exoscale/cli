package model

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"sync/atomic"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

type modelListTestServer struct {
	server        *httptest.Server
	models        []v3.ListModelsResponseEntry
	zoneListCount atomic.Int32
}

func newModelListTestServer(t *testing.T) *modelListTestServer {
	ts := &modelListTestServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/model", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(t, w, http.StatusOK, v3.ListModelsResponse{Models: ts.models})
		case http.MethodPost:
			writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-model-create"), State: v3.OperationStateSuccess})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/ai/model/", func(w http.ResponseWriter, r *http.Request) {
		id := path.Base(r.URL.Path)
		if r.Method == http.MethodGet {
			for _, m := range ts.models {
				if string(m.ID) == id {
					writeJSON(t, w, http.StatusOK, v3.GetModelResponse{ID: m.ID, Name: m.Name, State: v3.GetModelResponseState(m.State), ModelSize: m.ModelSize, CreatedAT: m.CreatedAT, UpdatedAT: m.UpdatedAT})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodDelete {
			writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-model-delete"), State: v3.OperationStateSuccess})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		ts.zoneListCount.Add(1)
		writeJSON(t, w, http.StatusOK, v3.ListZonesResponse{Zones: []v3.Zone{{APIEndpoint: v3.Endpoint(ts.server.URL), Name: v3.ZoneName("test-zone")}}})
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func setupModelList(t *testing.T, ts *modelListTestServer) func() {
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(ts.server.URL))
	return func() { ts.server.Close() }
}

func runModelListTest(t *testing.T, zoneFilter v3.ZoneName) (stdout, stderr string, err error) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := &ModelListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Zone:               zoneFilter,
	}
	err = runModelList(cmd, &outBuf, &errBuf)
	return outBuf.String(), errBuf.String(), err
}

func TestModelList(t *testing.T) {
	ts := newModelListTestServer(t)
	defer setupModelList(t, ts)()
	now := time.Now()
	ts.models = []v3.ListModelsResponseEntry{
		{ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1", State: v3.ListModelsResponseEntryStateReady, ModelSize: 0, CreatedAT: now, UpdatedAT: now},
		{ID: v3.UUID("22222222-2222-2222-2222-222222222222"), Name: "m2", State: v3.ListModelsResponseEntryStateCreating, ModelSize: 1024 * 1024 * 1024, CreatedAT: now, UpdatedAT: now},
	}
	defer withFormat(t, "json")()

	stdout, _, err := runModelListTest(t, "")
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
	defer setupModelList(t, ts)()
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
	defer setupModelList(t, ts)()
	defer withFormat(t, "json")()
	ts.models = nil

	stdout, _, err := runModelListTest(t, "")
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

// withFormat sets globalstate.OutputFormat for the duration of a test.
func withFormat(t *testing.T, f string) func() {
	t.Helper()
	prev := globalstate.OutputFormat
	globalstate.OutputFormat = f
	return func() { globalstate.OutputFormat = prev }
}
