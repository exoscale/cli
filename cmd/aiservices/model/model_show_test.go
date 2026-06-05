package model

import (
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
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
			testutils.WriteJSON(t, w, http.StatusOK, resp)
		case http.MethodPost:
			op := v3.Operation{ID: v3.UUID("op-model-create"), State: v3.OperationStateSuccess}
			testutils.WriteJSON(t, w, http.StatusOK, op)
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
						State:     v3.GetModelResponseState(m.State),
						ModelSize: m.ModelSize,
						CreatedAT: m.CreatedAT,
						UpdatedAT: m.UpdatedAT,
					}
					testutils.WriteJSON(t, w, http.StatusOK, resp)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodDelete:
			op := v3.Operation{ID: v3.UUID("op-model-delete"), State: v3.OperationStateSuccess}
			testutils.WriteJSON(t, w, http.StatusOK, op)
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
		testutils.WriteJSON(t, w, http.StatusOK, op)
	})
	mux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp := v3.ListZonesResponse{Zones: []v3.Zone{{APIEndpoint: v3.Endpoint(ts.server.URL), Name: v3.ZoneName("test-zone")}}}
		testutils.WriteJSON(t, w, http.StatusOK, resp)
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func TestModelShow(t *testing.T) {
	ts := newModelTestServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)
	now := time.Now()
	ts.models = []v3.ListModelsResponseEntry{{
		ID:        v3.UUID("11111111-1111-1111-1111-111111111111"),
		Name:      "m1",
		State:     v3.ListModelsResponseEntryStateReady,
		ModelSize: 1024 * 1024 * 1024 * 2,
		CreatedAT: now,
		UpdatedAT: now,
	}}

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
	if string(got.ID) != "11111111-1111-1111-1111-111111111111" || got.Name != "m1" || got.State != v3.GetModelResponseStateReady {
		t.Fatalf("unexpected model show output: %+v", got)
	}
	if got.ModelSize != "2.0 GiB" {
		t.Errorf("expected model size 2.0 GiB, got %q", got.ModelSize)
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
