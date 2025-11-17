package model

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

func newModelCreateServer(t *testing.T) *httptest.Server {
    mux := http.NewServeMux()
    mux.HandleFunc("/ai/model", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-model-create"), State: v3.OperationStateSuccess})
            return
        }
        w.WriteHeader(http.StatusMethodNotAllowed)
    })
    mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet { w.WriteHeader(http.StatusMethodNotAllowed); return }
        writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-model-create"), State: v3.OperationStateSuccess})
    })
    return httptest.NewServer(mux)
}

func TestModelCreateSuccessAndMissingName(t *testing.T) {
    srv := newModelCreateServer(t)
    defer srv.Close()
    exocmd.GContext = context.Background()
    globalstate.Quiet = true
    creds := credentials.NewStaticCredentials("key", "secret")
    client, err := v3.NewClient(creds)
    if err != nil { t.Fatalf("new client: %v", err) }
    globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(srv.URL))

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
