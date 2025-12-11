package deployment

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

type depActionsServer struct {
	server        *httptest.Server
	deployments   []v3.ListDeploymentsResponseEntry
	lastScaleBody string
}

func newDepActionsServer(t *testing.T) *depActionsServer {
	ts := &depActionsServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/deployment", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(t, w, http.StatusOK, v3.ListDeploymentsResponse{Deployments: ts.deployments})
		case http.MethodPost:
			writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-deploy-create"), State: v3.OperationStateSuccess})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/ai/deployment/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/ai/deployment/")
		parts := strings.Split(p, "/")
		id := parts[0]
		if id == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodGet {
			for _, d := range ts.deployments {
				if string(d.ID) == id {
					resp := v3.GetDeploymentResponse{ID: d.ID, Name: d.Name, Status: v3.GetDeploymentResponseStatus(d.Status), GpuType: d.GpuType, GpuCount: d.GpuCount, Replicas: d.Replicas, ServiceLevel: d.ServiceLevel, DeploymentURL: d.DeploymentURL, Model: d.Model, CreatedAT: d.CreatedAT, UpdatedAT: d.UpdatedAT}
					writeJSON(t, w, http.StatusOK, resp)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodDelete {
			writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-deploy-delete"), State: v3.OperationStateSuccess})
			return
		}
		if len(parts) == 2 && parts[1] == "scale" && r.Method == http.MethodPost {
			b, _ := io.ReadAll(r.Body)
			r.Body.Close()
			ts.lastScaleBody = string(b)
			writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-deploy-scale"), State: v3.OperationStateSuccess})
			return
		}
		if len(parts) == 2 && parts[1] == "api-key" && r.Method == http.MethodGet {
			writeJSON(t, w, http.StatusOK, v3.RevealDeploymentAPIKeyResponse{APIKey: "secret"})
			return
		}
		if len(parts) == 2 && parts[1] == "logs" && r.Method == http.MethodGet {
			writeJSON(t, w, http.StatusOK, v3.GetDeploymentLogsResponse("l1\nl2"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID(path.Base(r.URL.Path)), State: v3.OperationStateSuccess})
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func TestDeploymentDeleteScaleRevealLogs(t *testing.T) {
	ts := newDepActionsServer(t)
	defer ts.server.Close()
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(ts.server.URL))

	now := time.Now()
	ts.deployments = []v3.ListDeploymentsResponseEntry{{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", CreatedAT: now, UpdatedAT: now}}
	// delete by name
	del := &DeploymentDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployments: []string{"alpha"}, Force: true}
	if err := del.CmdRun(nil, nil); err != nil {
		t.Fatalf("delete: %v", err)
	}
	// delete multiple (add another deployment first)
	ts.deployments = append(ts.deployments, v3.ListDeploymentsResponseEntry{ID: v3.UUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), Name: "beta", CreatedAT: now, UpdatedAT: now})
	delMultiple := &DeploymentDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployments: []string{"alpha", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"}, Force: true}
	if err := delMultiple.CmdRun(nil, nil); err != nil {
		t.Fatalf("delete multiple: %v", err)
	}
	// scale by id
	sc := &DeploymentScaleCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", Size: 3}
	if err := sc.CmdRun(nil, nil); err != nil {
		t.Fatalf("scale: %v", err)
	}
	// reveal api key
	reveal := &DeploymentRevealAPIKeyCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "alpha"}
	var got string
	reveal.OutputFunc = func(o output.Outputter, err error) error {
		if err != nil {
			return err
		}
		out := o.(*DeploymentRevealAPIKeyOutput)
		got = out.APIKey
		return nil
	}
	if err := reveal.CmdRun(nil, nil); err != nil {
		t.Fatalf("reveal: %v", err)
	}
	if got != "secret" {
		t.Fatalf("unexpected api key: %s", got)
	}
	// logs
	logs := &DeploymentLogsCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "alpha"}
	if err := logs.CmdRun(nil, nil); err != nil {
		t.Fatalf("logs: %v", err)
	}
}

func TestDeploymentScaleZeroIncludesReplicas(t *testing.T) {
	ts := newDepActionsServer(t)
	defer ts.server.Close()
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(ts.server.URL))
	now := time.Now()
	ts.deployments = []v3.ListDeploymentsResponseEntry{{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", CreatedAT: now, UpdatedAT: now}}
	sc := &DeploymentScaleCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "alpha", Size: 0}
	if err := sc.CmdRun(nil, nil); err != nil {
		t.Fatalf("scale zero: %v", err)
	}
	if ts.lastScaleBody == "" {
		t.Fatalf("expected scale request body to be captured")
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(ts.lastScaleBody), &body); err != nil {
		t.Fatalf("invalid json body captured: %v", err)
	}
	v, ok := body["replicas"]
	if !ok {
		t.Fatalf("expected 'replicas' field in request body, got: %s", ts.lastScaleBody)
	}
	n, ok := v.(float64)
	if !ok || n != 0 {
		t.Fatalf("expected replicas to be 0, got: %v (body: %s)", v, ts.lastScaleBody)
	}
}
