package deployment

import (
    "context"
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
        if r.Method != http.MethodGet { w.WriteHeader(http.StatusMethodNotAllowed); return }
        writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-deploy-create"), State: v3.OperationStateSuccess})
    })
    srv := httptest.NewServer(mux)
    defer srv.Close()

    exocmd.GContext = context.Background()
    globalstate.Quiet = true
    creds := credentials.NewStaticCredentials("key", "secret")
    client, err := v3.NewClient(creds)
    if err != nil { t.Fatalf("new client: %v", err) }
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
