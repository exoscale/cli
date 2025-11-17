package model

import (
    "context"
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
                    writeJSON(t, w, http.StatusOK, v3.GetModelResponse{ID: m.ID, Name: m.Name, Status: v3.GetModelResponseStatus(m.Status), ModelSize: m.ModelSize, CreatedAT: m.CreatedAT, UpdatedAT: m.UpdatedAT})
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

func TestModelList(t *testing.T) {
    ts := newModelListTestServer(t)
    defer setupModelList(t, ts)()
    now := time.Now()
    ts.models = []v3.ListModelsResponseEntry{
        {ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1", Status: v3.ListModelsResponseEntryStatusReady, ModelSize: 0, CreatedAT: now, UpdatedAT: now},
        {ID: v3.UUID("22222222-2222-2222-2222-222222222222"), Name: "m2", Status: v3.ListModelsResponseEntryStatusCreating, ModelSize: 1234, CreatedAT: now, UpdatedAT: now},
    }
    cmd := &ModelListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
    if err := cmd.CmdRun(nil, nil); err != nil {
        t.Fatalf("model list: %v", err)
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
