package apikey

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

func newAIAPIKeyCreateServer(t *testing.T, captured *v3.CreateAIAPIKeyRequest) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/api-key", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		if err := json.Unmarshal(body, captured); err != nil {
			t.Fatalf("failed to unmarshal request: %v", err)
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.CreateAIAPIKeyResponse{
			ID:             v3.UUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			Name:           "my-key",
			Value:          "exo_ai_secret_value",
			Models:         &v3.AIAPIKeyModels{},
			Deployments:    &v3.AIAPIKeyDeployments{{ID: v3.UUID("cccccccc-cccc-cccc-cccc-cccccccccccc")}},
			AllModels:      boolPtr(true),
			AllDeployments: boolPtr(true),
		})
	})
	return httptest.NewServer(mux)
}

func boolPtr(b bool) *bool { return &b }

func TestAIAPIKeyCreate(t *testing.T) {
	var capturedRequest v3.CreateAIAPIKeyRequest
	srv := newAIAPIKeyCreateServer(t, &capturedRequest)
	defer srv.Close()
	testutils.SetupV3Client(t, srv.URL)

	var got *AIAPIKeyCreateOutput
	c := &AIAPIKeyCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Name:               "my-key",
		AllModels:          true,
		AllDeployments:     true,
	}
	c.OutputFunc = func(o output.Outputter, err error) error {
		if err != nil {
			return err
		}
		got = o.(*AIAPIKeyCreateOutput)
		return nil
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("api-key create: %v", err)
	}

	if capturedRequest.Name != "my-key" {
		t.Errorf("expected name %q in request, got %q", "my-key", capturedRequest.Name)
	}
	if capturedRequest.AllModels == nil || !*capturedRequest.AllModels {
		t.Errorf("expected all-models in request, got %v", capturedRequest.AllModels)
	}
	if capturedRequest.Models != nil {
		t.Errorf("expected models omitted from request, got %v", *capturedRequest.Models)
	}
	if capturedRequest.AllDeployments == nil || !*capturedRequest.AllDeployments {
		t.Errorf("expected all-deployments in request, got %v", capturedRequest.AllDeployments)
	}
	if capturedRequest.Deployments != nil {
		t.Errorf("expected deployments omitted from request, got %v", *capturedRequest.Deployments)
	}

	if got == nil {
		t.Fatal("no output captured")
	}
	if got.Value != "exo_ai_secret_value" {
		t.Errorf("expected secret value to be surfaced, got %q", got.Value)
	}
	if got.ID != v3.UUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb") {
		t.Errorf("unexpected id: %q", got.ID)
	}
	if !got.AllModels || !got.AllDeployments {
		t.Errorf("expected all-models/all-deployments true in output, got %+v", got)
	}
	if got.Deployments == nil || len(got.Deployments) != 1 || got.Deployments[0] != "cccccccc-cccc-cccc-cccc-cccccccccccc" {
		t.Errorf("expected output deployments to flatten {id} refs, got %v", got.Deployments)
	}
}

func TestAIAPIKeyCreateAllowlists(t *testing.T) {
	var capturedRequest v3.CreateAIAPIKeyRequest
	srv := newAIAPIKeyCreateServer(t, &capturedRequest)
	defer srv.Close()
	testutils.SetupV3Client(t, srv.URL)

	c := &AIAPIKeyCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Name:               "my-key",
		Models:             []string{"llama-3-8b"},
		Deployments:        []string{"cccccccc-cccc-cccc-cccc-cccccccccccc"},
	}
	c.OutputFunc = func(output.Outputter, error) error { return nil }
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("api-key create with allowlists: %v", err)
	}

	if capturedRequest.AllModels != nil || capturedRequest.AllDeployments != nil {
		t.Errorf("expected all-* flags omitted from request, got %v/%v", capturedRequest.AllModels, capturedRequest.AllDeployments)
	}
	if capturedRequest.Models == nil || len(*capturedRequest.Models) != 1 || (*capturedRequest.Models)[0] != "llama-3-8b" {
		t.Errorf("expected request models [llama-3-8b], got %v", capturedRequest.Models)
	}
	if capturedRequest.Deployments == nil || len(*capturedRequest.Deployments) != 1 || (*capturedRequest.Deployments)[0].ID != "cccccccc-cccc-cccc-cccc-cccccccccccc" {
		t.Errorf("expected deployment ref in request, got %v", capturedRequest.Deployments)
	}
}

func TestAIAPIKeyCreateConflictingFlags(t *testing.T) {
	for _, c := range []*AIAPIKeyCreateCmd{
		{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Name: "k", AllModels: true, Models: []string{"llama-3-8b"}},
		{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Name: "k", AllDeployments: true, Deployments: []string{"cccccccc-cccc-cccc-cccc-cccccccccccc"}},
	} {
		if err := c.CmdRun(nil, nil); err == nil {
			t.Errorf("expected conflict error for %+v", c)
		}
	}
}

func TestAIAPIKeyCreateInvalidDeploymentID(t *testing.T) {
	c := &AIAPIKeyCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Name:               "my-key",
		Deployments:        []string{"not-a-uuid"},
	}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatal("expected error for non-UUID deployment ID")
	}
}

func TestAIAPIKeyCreateOmitsEmptyAccessLists(t *testing.T) {
	var capturedRequest v3.CreateAIAPIKeyRequest
	srv := newAIAPIKeyCreateServer(t, &capturedRequest)
	defer srv.Close()
	testutils.SetupV3Client(t, srv.URL)

	c := &AIAPIKeyCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Name:               "my-key",
	}
	c.OutputFunc = func(output.Outputter, error) error { return nil }
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("api-key create: %v", err)
	}

	if capturedRequest.Models != nil {
		t.Errorf("expected models omitted from request, got %v", *capturedRequest.Models)
	}
	if capturedRequest.Deployments != nil {
		t.Errorf("expected deployments omitted from request, got %v", *capturedRequest.Deployments)
	}
	if capturedRequest.AllModels != nil || capturedRequest.AllDeployments != nil {
		t.Errorf("expected all-* flags omitted from request, got %v/%v", capturedRequest.AllModels, capturedRequest.AllDeployments)
	}
}

func TestAIAPIKeyCreateWithoutName(t *testing.T) {
	c := &AIAPIKeyCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestAIAPIKeyCreateWithZone(t *testing.T) {
	var capturedRequest v3.CreateAIAPIKeyRequest
	zoneSrv := newAIAPIKeyCreateServer(t, &capturedRequest)
	defer zoneSrv.Close()

	controlMux := http.NewServeMux()
	controlMux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListZonesResponse{
			Zones: []v3.Zone{{Name: v3.ZoneName("z1"), APIEndpoint: v3.Endpoint(zoneSrv.URL)}},
		})
	})
	controlSrv := httptest.NewServer(controlMux)
	defer controlSrv.Close()
	testutils.SetupV3Client(t, controlSrv.URL)

	var got *AIAPIKeyCreateOutput
	c := &AIAPIKeyCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Name:               "my-key",
		Zone:               "z1",
	}
	c.OutputFunc = func(o output.Outputter, err error) error {
		if err != nil {
			return err
		}
		got = o.(*AIAPIKeyCreateOutput)
		return nil
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("api-key create with zone: %v", err)
	}

	if capturedRequest.Name != "my-key" {
		t.Fatalf("expected the create request to hit the zone endpoint, got %+v", capturedRequest)
	}
	if got == nil || got.ID != "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" {
		t.Fatalf("unexpected output: %+v", got)
	}
}
