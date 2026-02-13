package deployment

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

func TestDeploymentUpdate(t *testing.T) {
	var capturedRequest v3.UpdateDeploymentRequest
	var capturedID string

	mux := http.NewServeMux()
	// Mock ListDeployments for resolution
	mux.HandleFunc("/ai/deployment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := v3.ListDeploymentsResponse{
				Deployments: []v3.ListDeploymentsResponseEntry{
					{
						ID:   v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
						Name: "alpha",
					},
				},
			}
			writeJSON(t, w, http.StatusOK, resp)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Mock UpdateDeployment
	mux.HandleFunc("/ai/deployment/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			capturedID = path.Base(r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			r.Body.Close()
			if err := json.Unmarshal(body, &capturedRequest); err != nil {
				t.Fatalf("failed to unmarshal request: %v", err)
			}
			writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-update"), State: v3.OperationStateSuccess})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-update"), State: v3.OperationStateSuccess})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(srv.URL))

	c := &DeploymentUpdateCmd{
		CliCommandSettings:        exocmd.DefaultCLICmdSettings(),
		Deployment:                "alpha",
		Name:                      "new-name",
		InferenceEngineVersion:    "0.15.1",
		InferenceEngineParameters: "--foo --bar",
	}

	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment update: %v", err)
	}

	if capturedID != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Errorf("expected ID %q, got %q", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", capturedID)
	}
	if capturedRequest.Name != "new-name" {
		t.Errorf("expected name %q, got %q", "new-name", capturedRequest.Name)
	}
	if string(capturedRequest.InferenceEngineVersion) != "0.15.1" {
		t.Errorf("expected version %q, got %q", "0.15.1", capturedRequest.InferenceEngineVersion)
	}
	expectedParams := []string{"--foo", "--bar"}
	if len(capturedRequest.InferenceEngineParameters) != len(expectedParams) {
		t.Fatalf("expected %d params, got %d", len(expectedParams), len(capturedRequest.InferenceEngineParameters))
	}
	for i, p := range expectedParams {
		if capturedRequest.InferenceEngineParameters[i] != p {
			t.Errorf("param[%d] expected %q, got %q", i, p, capturedRequest.InferenceEngineParameters[i])
		}
	}
}
