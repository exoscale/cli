package deployment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

type depShowServer struct {
	server      *httptest.Server
	deployments []v3.ListDeploymentsResponseEntry
}

func newDepShowServer(t *testing.T) *depShowServer {
	ts := &depShowServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/deployment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeJSON(t, w, http.StatusOK, v3.ListDeploymentsResponse{Deployments: ts.deployments})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/ai/deployment/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := path.Base(r.URL.Path)
		for _, d := range ts.deployments {
			if string(d.ID) == id {
				resp := v3.GetDeploymentResponse{
					ID:           d.ID,
					Name:         d.Name,
					State:        v3.GetDeploymentResponseState(d.State),
					GpuType:      d.GpuType,
					GpuCount:     d.GpuCount,
					Replicas:     d.Replicas,
					ServiceLevel: d.ServiceLevel,
					// InferenceEngineVersion is not in ListDeploymentsResponseEntry so we hardcode it for show tests if needed,
					// or just leave it empty. GetDeploymentResponse DOES have it.
					InferenceEngineVersion: "0.15.1",
					DeploymentURL:          d.DeploymentURL,
					Model:                  d.Model,
					CreatedAT:              d.CreatedAT,
					UpdatedAT:              d.UpdatedAT,
				}
				writeJSON(t, w, http.StatusOK, resp)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func TestDeploymentShowByIDAndName(t *testing.T) {
	ts := newDepShowServer(t)
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
	ts.deployments = []v3.ListDeploymentsResponseEntry{{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", State: v3.ListDeploymentsResponseEntryStateReady, GpuType: "gpua5000", GpuCount: 1, Replicas: 1, ServiceLevel: "pro", DeploymentURL: "https://u", Model: &v3.ModelRef{ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1"}, CreatedAT: now, UpdatedAT: now}}

	// by ID
	cmd := &DeploymentShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	var got DeploymentShowOutput
	cmd.OutputFunc = func(o output.Outputter, err error) error {
		if err != nil {
			return err
		}
		got = *(o.(*DeploymentShowOutput))
		return nil
	}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment show by id: %v", err)
	}
	if got.Name != "alpha" || got.GPUType != "gpua5000" || got.State != v3.GetDeploymentResponseStateReady {
		t.Fatalf("unexpected show output: %+v", got)
	}
	if got.InferenceEngineVersion != "0.15.1" {
		t.Errorf("expected inference engine version 0.15.1, got %q", got.InferenceEngineVersion)
	}
	// by name
	cmd = &DeploymentShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "alpha"}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment show by name: %v", err)
	}
}
