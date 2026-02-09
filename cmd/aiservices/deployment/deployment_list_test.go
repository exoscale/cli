package deployment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

type depListServer struct {
	server      *httptest.Server
	deployments []v3.ListDeploymentsResponseEntry
	zones       int
}

func newDepListServer(t *testing.T) *depListServer {
	ts := &depListServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/deployment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeJSON(t, w, http.StatusOK, v3.ListDeploymentsResponse{Deployments: ts.deployments})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		ts.zones++
		writeJSON(t, w, http.StatusOK, v3.ListZonesResponse{Zones: []v3.Zone{{APIEndpoint: v3.Endpoint(ts.server.URL), Name: v3.ZoneName("test-zone")}}})
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func depSetup(t *testing.T, url string) func() {
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(url))
	return func() {}
}

func writeJSON(t *testing.T, w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

func TestDeploymentList(t *testing.T) {
	ts := newDepListServer(t)
	defer ts.server.Close()
	defer depSetup(t, ts.server.URL)()
	now := time.Now()
	ts.deployments = []v3.ListDeploymentsResponseEntry{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "d1", Status: v3.ListDeploymentsResponseEntryStatusReady, GpuType: "gpua5000", GpuCount: 1, Replicas: 2, ServiceLevel: "pro", DeploymentURL: "https://u", Model: &v3.ModelRef{ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1"}, CreatedAT: now, UpdatedAT: now},
		{ID: v3.UUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), Name: "d2", Status: v3.ListDeploymentsResponseEntryStatusCreating, GpuType: "gpua5000", GpuCount: 2, Replicas: 1, ServiceLevel: "pro", DeploymentURL: "", Model: nil, CreatedAT: now, UpdatedAT: now},
	}
	cmd := &DeploymentListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	cmd.OutputFunc = func(out output.Outputter, err error) error {
		if err != nil {
			return err
		}
		o := out.(*DeploymentListOutput)
		if len(*o) != 2 {
			t.Fatalf("expected 2 deployments, got %d", len(*o))
		}
		for _, d := range *o {
			if d.Zone != "test-zone" {
				t.Errorf("expected zone %q, got %q", "test-zone", d.Zone)
			}
		}
		return nil
	}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment list: %v", err)
	}
}

func TestDeploymentListUsesZone(t *testing.T) {
	ts := newDepListServer(t)
	defer ts.server.Close()
	defer depSetup(t, ts.server.URL)()
	cmd := &DeploymentListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Zone: v3.ZoneName("test-zone")}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment list: %v", err)
	}
	if ts.zones == 0 {
		t.Fatalf("expected zone list endpoint to be called")
	}
}
