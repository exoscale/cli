package deployment

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

func TestDeploymentCreateValidationAndSuccess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/deployment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-deploy-create"), State: v3.OperationStateSuccess})
	})
	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-deploy-create"), State: v3.OperationStateSuccess})
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

	// missing gpu flags
	c := &DeploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatalf("expected error for missing gpu flags")
	}
	// missing model flags
	c = &DeploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), GPUType: "gpua5000", GPUCount: 1}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatalf("expected error for missing model flags")
	}
	// invalid model id
	c = &DeploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), GPUType: "gpua5000", GPUCount: 1, ModelID: "bad"}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatalf("expected invalid model id error")
	}
	// success with model name
	c = &DeploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Name: "dep", GPUType: "gpua5000", GPUCount: 1, Replicas: 1, ModelName: "m1"}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment create: %v", err)
	}
}

func TestDeploymentCreateWithInferenceEngineParameters(t *testing.T) {
	var capturedRequest v3.CreateDeploymentRequest
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/deployment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		if err := json.Unmarshal(body, &capturedRequest); err != nil {
			t.Fatalf("failed to unmarshal request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-deploy-create"), State: v3.OperationStateSuccess})
	})
	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-deploy-create"), State: v3.OperationStateSuccess})
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

	// Test with space-separated inference engine parameters
	c := &DeploymentCreateCmd{
		CliCommandSettings:        exocmd.DefaultCLICmdSettings(),
		Name:                      "dep-with-params",
		GPUType:                   "gpua5000",
		GPUCount:                  1,
		Replicas:                  1,
		ModelName:                 "m1",
		InferenceEngineParameters: "--gpu-memory-usage=0.8 --max-tokens=4096 --disable-x-feature",
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment create with params: %v", err)
	}

	// Verify the parameters were parsed correctly
	expectedParams := []string{"--gpu-memory-usage=0.8", "--max-tokens=4096", "--disable-x-feature"}
	if len(capturedRequest.InferenceEngineParameters) != len(expectedParams) {
		t.Fatalf("expected %d parameters, got %d", len(expectedParams), len(capturedRequest.InferenceEngineParameters))
	}
	for i, param := range expectedParams {
		if capturedRequest.InferenceEngineParameters[i] != param {
			t.Errorf("parameter[%d]: expected %q, got %q", i, param, capturedRequest.InferenceEngineParameters[i])
		}
	}
}

func TestDeploymentCreateInferenceEngineHelp(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/deployment/inference-engine-help", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp := v3.GetInferenceEngineHelpResponse{
			Parameters: []v3.InferenceEngineParameterEntry{
				{
					Name:        "config-format",
					Flags:       []string{"--config-format"},
					Type:        "enum",
					Default:     "auto",
					Section:     "ModelConfig",
					Description: "The format of the model config to load.",
				},
				{
					Name:        "max-model-len",
					Flags:       []string{"--max-model-len"},
					Type:        "integer",
					Default:     "None",
					Section:     "ModelConfig",
					Description: "Model context length.",
				},
			},
		}
		writeJSON(t, w, http.StatusOK, resp)
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

	c := &DeploymentCreateCmd{
		CliCommandSettings:  exocmd.DefaultCLICmdSettings(),
		InferenceEngineHelp: true,
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment create help without name: %v", err)
	}

	c = &DeploymentCreateCmd{
		CliCommandSettings:  exocmd.DefaultCLICmdSettings(),
		Name:                "test-deploy",
		InferenceEngineHelp: true,
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment create help with name: %v", err)
	}
}
