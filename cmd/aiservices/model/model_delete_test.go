package model

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

func newModelDeleteServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/model/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-model-delete"), State: v3.OperationStateSuccess})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-model-delete"), State: v3.OperationStateSuccess})
	})
	return httptest.NewServer(mux)
}

func TestModelDeleteInvalidUUIDAndSuccess(t *testing.T) {
	srv := newModelDeleteServer(t)
	defer srv.Close()
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(srv.URL))

	// invalid UUID
	cmd := &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), ID: "not-a-uuid", Force: true}
	if err := cmd.CmdRun(nil, nil); err == nil || !regexp.MustCompile(`invalid model ID`).MatchString(err.Error()) {
		t.Fatalf("expected invalid uuid error, got %v", err)
	}
	// success
	cmd = &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), ID: "33333333-3333-3333-3333-333333333333", Force: true}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model delete: %v", err)
	}
}
