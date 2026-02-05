package model

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

// Minimal test server + helpers used across model tests in this package.
type modelTestServer struct {
	server *httptest.Server
	models []v3.ListModelsResponseEntry
}

func newModelTestServer(t *testing.T) *modelTestServer {
	ts := &modelTestServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/model", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := v3.ListModelsResponse{Models: ts.models}
			writeJSON(t, w, http.StatusOK, resp)
		case http.MethodPost:
			op := v3.Operation{ID: v3.UUID("op-model-create"), State: v3.OperationStateSuccess}
			writeJSON(t, w, http.StatusOK, op)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/ai/model/", func(w http.ResponseWriter, r *http.Request) {
		id := path.Base(r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			for _, m := range ts.models {
				if string(m.ID) == id {
					resp := v3.GetModelResponse{
						ID:        m.ID,
						Name:      m.Name,
						Status:    v3.GetModelResponseStatus(m.Status),
						ModelSize: m.ModelSize,
						CreatedAT: m.CreatedAT,
						UpdatedAT: m.UpdatedAT,
					}
					writeJSON(t, w, http.StatusOK, resp)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodDelete:
			op := v3.Operation{ID: v3.UUID("op-model-delete"), State: v3.OperationStateSuccess}
			writeJSON(t, w, http.StatusOK, op)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		op := v3.Operation{ID: v3.UUID(path.Base(r.URL.Path)), State: v3.OperationStateSuccess}
		writeJSON(t, w, http.StatusOK, op)
	})
	mux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp := v3.ListZonesResponse{Zones: []v3.Zone{{APIEndpoint: v3.Endpoint(ts.server.URL), Name: v3.ZoneName("test-zone")}}}
		writeJSON(t, w, http.StatusOK, resp)
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func writeJSON(t *testing.T, w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

func newV3Client(t *testing.T, endpoint string) *v3.Client {
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return client.WithEndpoint(v3.Endpoint(endpoint))
}

func modelSetup(t *testing.T, ts *modelTestServer) func() {
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	globalstate.EgoscaleV3Client = newV3Client(t, ts.server.URL)
	return func() { ts.server.Close() }
}

func TestModelShow(t *testing.T) {
	ts := newModelTestServer(t)
	defer modelSetup(t, ts)()
	now := time.Now()
	ts.models = []v3.ListModelsResponseEntry{{ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1", Status: v3.ListModelsResponseEntryStatusReady, ModelSize: 123, CreatedAT: now, UpdatedAT: now}}

	cmd := &ModelShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Model: "11111111-1111-1111-1111-111111111111"}
	var got ModelShowOutput
	cmd.OutputFunc = func(o output.Outputter, err error) error {
		if err != nil {
			return err
		}
		got = *(o.(*ModelShowOutput))
		return nil
	}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model show: %v", err)
	}
	if string(got.ID) != "11111111-1111-1111-1111-111111111111" || got.Name != "m1" || got.Status != v3.GetModelResponseStatusReady {
		t.Fatalf("unexpected model show output: %+v", got)
	}

	// Test show by name
	cmd = &ModelShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Model: "m1"}
	cmd.OutputFunc = func(o output.Outputter, err error) error {
		if err != nil {
			return err
		}
		got = *(o.(*ModelShowOutput))
		return nil
	}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model show by name: %v", err)
	}
	if string(got.ID) != "11111111-1111-1111-1111-111111111111" || got.Name != "m1" {
		t.Fatalf("unexpected model show output (by name): %+v", got)
	}
}
