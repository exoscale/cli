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

	// invalid UUID without force
	cmd := &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), IDs: []string{"not-a-uuid"}, Force: false}
	if err := cmd.CmdRun(nil, nil); err == nil || !regexp.MustCompile(`invalid model ID`).MatchString(err.Error()) {
		t.Fatalf("expected invalid uuid error, got %v", err)
	}
	// invalid UUID with force (should skip with warning, no error)
	cmd = &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), IDs: []string{"not-a-uuid"}, Force: true}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("expected no error with force flag, got %v", err)
	}
	// success
	cmd = &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), IDs: []string{"33333333-3333-3333-3333-333333333333"}, Force: true}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model delete: %v", err)
	}
	// multiple models
	cmd = &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), IDs: []string{"33333333-3333-3333-3333-333333333333", "44444444-4444-4444-4444-444444444444"}, Force: true}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model delete multiple: %v", err)
	}
}

func TestModelDeleteCmd_CmdAliases(t *testing.T) {
	cmd := &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	aliases := cmd.CmdAliases()
	if len(aliases) == 0 {
		t.Fatal("CmdAliases() returned empty slice")
	}
	// Verify it returns the standard delete aliases
	expectedAliases := exocmd.GDeleteAlias
	if len(aliases) != len(expectedAliases) {
		t.Fatalf("expected %d aliases, got %d", len(expectedAliases), len(aliases))
	}
	for i, alias := range aliases {
		if alias != expectedAliases[i] {
			t.Fatalf("expected alias[%d] to be %q, got %q", i, expectedAliases[i], alias)
		}
	}
}
