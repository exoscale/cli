package deployment

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

type depHelperServer struct {
	server      *httptest.Server
	deployments []v3.ListDeploymentsResponseEntry
}

func newDepHelperServer(t *testing.T) *depHelperServer {
	ts := &depHelperServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/deployment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeJSON(t, w, http.StatusOK, v3.ListDeploymentsResponse{Deployments: ts.deployments})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/ai/deployment/", func(w http.ResponseWriter, r *http.Request) {
		id := path.Base(r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			for _, d := range ts.deployments {
				if string(d.ID) == id {
					writeJSON(t, w, http.StatusOK, v3.GetDeploymentResponse{ID: d.ID, Name: d.Name, Status: v3.GetDeploymentResponseStatus(d.Status), GpuType: d.GpuType, GpuCount: d.GpuCount, Replicas: d.Replicas, ServiceLevel: d.ServiceLevel, DeploymentURL: d.DeploymentURL, Model: d.Model, CreatedAT: d.CreatedAT, UpdatedAT: d.UpdatedAT})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodDelete:
			writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op"), State: v3.OperationStateSuccess})
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, "/scale") {
				io.Copy(io.Discard, r.Body)
				r.Body.Close()
				writeJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op"), State: v3.OperationStateSuccess})
				return
			}
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func TestResolveDeploymentIDByIDAndName(t *testing.T) {
	ts := newDepHelperServer(t)
	defer ts.server.Close()
	now := time.Now()
	ts.deployments = []v3.ListDeploymentsResponseEntry{{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", CreatedAT: now, UpdatedAT: now}}
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client = client.WithEndpoint(v3.Endpoint(ts.server.URL))
	ctx := context.Background()

	// by ID
	id, err := ResolveDeploymentID(ctx, client, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	if err != nil || string(id) != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("resolve by id failed: %v %v", id, err)
	}
	// by name
	id, err = ResolveDeploymentID(ctx, client, "alpha")
	if err != nil || string(id) != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("resolve by name failed: %v %v", id, err)
	}
	// not found
	_, err = ResolveDeploymentID(ctx, client, "missing")
	if err == nil || !strings.Contains(err.Error(), "deployment \"missing\" not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}
