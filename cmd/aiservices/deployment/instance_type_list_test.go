package deployment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

type instanceTypeListServer struct {
	server        *httptest.Server
	instanceTypes []v3.InstanceTypeEntry
	zones         int
}

func newInstanceTypeListServer(t *testing.T) *instanceTypeListServer {
	ts := &instanceTypeListServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/instance-type", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeJSON(t, w, http.StatusOK, v3.ListAIInstanceTypesResponse{InstanceTypes: ts.instanceTypes})
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

func instanceTypeSetup(t *testing.T, url string) func() {
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

func TestInstanceTypeList(t *testing.T) {
	ts := newInstanceTypeListServer(t)
	defer ts.server.Close()
	defer instanceTypeSetup(t, ts.server.URL)()

	trueVal := true
	falseVal := false
	ts.instanceTypes = []v3.InstanceTypeEntry{
		{Family: "gpu-a5000", Authorized: &falseVal},
		{Family: "gpu-a100", Authorized: &trueVal},
	}

	cmd := &InstanceTypeListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	cmd.OutputFunc = func(out output.Outputter, err error) error {
		if err != nil {
			return err
		}
		o := out.(*InstanceTypeListOutput)
		if len(*o) != 2 {
			t.Fatalf("expected 2 instance types, got %d", len(*o))
		}
		if (*o)[0].Family != "gpu-a100" {
			t.Errorf("expected first family gpu-a100, got %s", (*o)[0].Family)
		}
		if (*o)[1].Family != "gpu-a5000" {
			t.Errorf("expected second family gpu-a5000, got %s", (*o)[1].Family)
		}
		return nil
	}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("instance type list: %v", err)
	}
}

func TestInstanceTypeListSortByZoneAndFamily(t *testing.T) {
	out := InstanceTypeListOutput{
		{Family: "gpu-a100", Authorized: true, Zone: "ch-gva-2"},
		{Family: "gpu-a5000", Authorized: true, Zone: "ch-dk-2"},
		{Family: "gpu-a100", Authorized: true, Zone: "ch-dk-2"},
		{Family: "gpu-h100", Authorized: true, Zone: "ch-gva-2"},
	}

	sortInstanceTypeListOutput(out)

	if out[0].Zone != "ch-dk-2" || out[0].Family != "gpu-a100" {
		t.Errorf("expected ch-dk-2/gpu-a100 first, got %s/%s", out[0].Zone, out[0].Family)
	}
	if out[1].Zone != "ch-dk-2" || out[1].Family != "gpu-a5000" {
		t.Errorf("expected ch-dk-2/gpu-a5000 second, got %s/%s", out[1].Zone, out[1].Family)
	}
	if out[2].Zone != "ch-gva-2" || out[2].Family != "gpu-a100" {
		t.Errorf("expected ch-gva-2/gpu-a100 third, got %s/%s", out[2].Zone, out[2].Family)
	}
	if out[3].Zone != "ch-gva-2" || out[3].Family != "gpu-h100" {
		t.Errorf("expected ch-gva-2/gpu-h100 fourth, got %s/%s", out[3].Zone, out[3].Family)
	}
}

func TestInstanceTypeListUsesZone(t *testing.T) {
	ts := newInstanceTypeListServer(t)
	defer ts.server.Close()
	defer instanceTypeSetup(t, ts.server.URL)()

	cmd := &InstanceTypeListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Zone: v3.ZoneName("test-zone")}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("instance type list: %v", err)
	}
	if ts.zones != 1 {
		t.Fatalf("expected zone list endpoint to be called exactly once, got %d", ts.zones)
	}
}

func TestInstanceTypeListOutputToJSON(t *testing.T) {
	trueVal := true
	out := InstanceTypeListOutput{
		{Family: "gpu-a100", Authorized: trueVal, Zone: "ch-gva-2"},
		{Family: "gpu-a5000", Authorized: false, Zone: "ch-gva-2"},
	}

	jsonBytes, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}

	expected := `[{"family":"gpu-a100","authorized":true,"zone":"ch-gva-2"},{"family":"gpu-a5000","authorized":false,"zone":"ch-gva-2"}]`
	if string(jsonBytes) != expected {
		t.Errorf("expected %s, got %s", expected, string(jsonBytes))
	}
}

func TestInstanceTypeListCmd_CmdShort(t *testing.T) {
	cmd := &InstanceTypeListCmd{}
	if cmd.CmdShort() != "List AI instance types" {
		t.Errorf("expected %q, got %q", "List AI instance types", cmd.CmdShort())
	}
}
