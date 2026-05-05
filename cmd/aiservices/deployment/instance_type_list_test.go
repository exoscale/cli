package deployment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
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

func instanceTypeSetup(t *testing.T, url string) {
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(url))
}

func TestInstanceTypeList(t *testing.T) {
	ts := newInstanceTypeListServer(t)
	defer ts.server.Close()
	instanceTypeSetup(t, ts.server.URL)

	trueVal := true
	falseVal := false
	ts.instanceTypes = []v3.InstanceTypeEntry{
		{Family: "gpu-a5000", Authorized: &falseVal},
		{Family: "gpu-a100", Authorized: &trueVal},
	}

	globalstate.OutputFormat = "json"
	defer func() { globalstate.OutputFormat = "" }()

	cmd := &InstanceTypeListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	var outBuf bytes.Buffer
	if err := runInstanceTypeList(cmd, &outBuf, nil); err != nil {
		t.Fatalf("instance type list: %v", err)
	}

	var rows []InstanceTypeListItemOutput
	if err := json.Unmarshal(outBuf.Bytes(), &rows); err != nil {
		t.Fatalf("invalid json: %v\nstdout: %s", err, outBuf.String())
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 instance types, got %d", len(rows))
	}
	for _, r := range rows {
		if r.Zone != "test-zone" {
			t.Errorf("expected zone test-zone, got %s", r.Zone)
		}
	}
}

func TestInstanceTypeListUsesZone(t *testing.T) {
	ts := newInstanceTypeListServer(t)
	defer ts.server.Close()
	instanceTypeSetup(t, ts.server.URL)

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
	out := []InstanceTypeListItemOutput{
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
