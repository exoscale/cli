package model

import (
	"net/http"
	"net/http/httptest"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

func newModelCreateServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/model", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-model-create"), State: v3.OperationStateSuccess})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-model-create"), State: v3.OperationStateSuccess})
	})
	return httptest.NewServer(mux)
}

func TestModelCreateSuccessAndMissingName(t *testing.T) {
	srv := newModelCreateServer(t)
	defer srv.Close()
	testutils.SetupV3Client(t, srv.URL)

	// missing name
	cmd := &ModelCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	if err := cmd.CmdRun(nil, nil); err == nil || err.Error() != "NAME is required" {
		t.Fatalf("expected name required error, got %v", err)
	}
	// success
	cmd = &ModelCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Name: "foo", HuggingfaceToken: "tok"}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model create: %v", err)
	}
}
