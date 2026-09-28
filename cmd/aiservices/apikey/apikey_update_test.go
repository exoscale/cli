package apikey

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

type aiAPIKeyUpdateServer struct {
	server *httptest.Server
	keys   []v3.ListAIAPIKeysResponseEntry

	captured *v3.UpdateAIAPIKeyRequest
}

func newAIAPIKeyUpdateServer(t *testing.T) *aiAPIKeyUpdateServer {
	t.Helper()
	ts := &aiAPIKeyUpdateServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/api-key", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListAIAPIKeysResponse{AIAPIKeys: ts.keys})
	})
	mux.HandleFunc("/ai/api-key/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := path.Base(r.URL.Path)
		for _, k := range ts.keys {
			if string(k.ID) != id {
				continue
			}
			body, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			if err := json.Unmarshal(body, ts.captured); err != nil {
				t.Fatalf("failed to unmarshal request: %v", err)
			}
			testutils.WriteJSON(t, w, http.StatusOK, v3.UpdateAIAPIKeyResponse{
				ID:             k.ID,
				Name:           k.Name,
				Models:         &v3.AIAPIKeyModels{},
				Deployments:    &v3.AIAPIKeyDeployments{{ID: v3.UUID("cccccccc-cccc-cccc-cccc-cccccccccccc")}},
				AllModels:      boolPtr(true),
				AllDeployments: boolPtr(true),
				CreatedAT:      k.CreatedAT,
				UpdatedAT:      k.UpdatedAT,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

// newAIAPIKeyUpdateCmd registers the update command so the generated flags
// exist, and marks the given flags as set, so that cmd.Flags().Changed()
// behaves as it does in production.
func newAIAPIKeyUpdateCmd(t *testing.T, c *AIAPIKeyUpdateCmd, changed ...string) *cobra.Command {
	t.Helper()

	parent := &cobra.Command{Use: "test"}
	if err := exocmd.RegisterCLICommand(parent, c); err != nil {
		t.Fatalf("register command: %v", err)
	}

	cmd := parent.Commands()[0]
	for _, name := range changed {
		if err := cmd.Flags().Set(name, cmd.Flags().Lookup(name).Value.String()); err != nil {
			t.Fatalf("set flag %s: %v", name, err)
		}
	}

	return cmd
}

func TestAIAPIKeyUpdate(t *testing.T) {
	now := time.Now()
	ts := newAIAPIKeyUpdateServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	ts.keys = []v3.ListAIAPIKeysResponseEntry{{
		ID:             v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:           "alpha",
		Models:         &v3.AIAPIKeyModels{},
		Deployments:    &v3.AIAPIKeyDeployments{},
		AllModels:      boolPtr(false),
		AllDeployments: boolPtr(false),
		CreatedAT:      now,
		UpdatedAT:      now,
	}}

	var capturedRequest v3.UpdateAIAPIKeyRequest
	ts.captured = &capturedRequest

	var got *AIAPIKeyUpdateOutput
	c := &AIAPIKeyUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Key:                "alpha",
		AllModels:          true,
		AllDeployments:     true,
	}
	c.OutputFunc = func(o output.Outputter, err error) error {
		if err != nil {
			return err
		}
		got = o.(*AIAPIKeyUpdateOutput)
		return nil
	}
	cmd := newAIAPIKeyUpdateCmd(t, c, "all-models", "all-deployments")
	if err := c.CmdRun(cmd, nil); err != nil {
		t.Fatalf("api-key update: %v", err)
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
	if got == nil || got.ID != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("unexpected update output: %+v", got)
	}
	if !got.AllModels || !got.AllDeployments {
		t.Errorf("expected all-models/all-deployments true in output, got %+v", got)
	}
	if got.Deployments == nil || len(got.Deployments) != 1 || got.Deployments[0] != "cccccccc-cccc-cccc-cccc-cccccccccccc" {
		t.Errorf("expected output deployments to flatten {id} refs, got %v", got.Deployments)
	}
}

func TestAIAPIKeyUpdateAllowlists(t *testing.T) {
	now := time.Now()
	ts := newAIAPIKeyUpdateServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	ts.keys = []v3.ListAIAPIKeysResponseEntry{{
		ID:             v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:           "alpha",
		Models:         &v3.AIAPIKeyModels{},
		Deployments:    &v3.AIAPIKeyDeployments{},
		AllModels:      boolPtr(false),
		AllDeployments: boolPtr(false),
		CreatedAT:      now,
		UpdatedAT:      now,
	}}

	var capturedRequest v3.UpdateAIAPIKeyRequest
	ts.captured = &capturedRequest

	c := &AIAPIKeyUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Key:                "alpha",
		Models:             []string{"llama-3-8b"},
		Deployments:        []string{"cccccccc-cccc-cccc-cccc-cccccccccccc"},
	}
	c.OutputFunc = func(output.Outputter, error) error { return nil }
	cmd := newAIAPIKeyUpdateCmd(t, c)
	if err := c.CmdRun(cmd, nil); err != nil {
		t.Fatalf("api-key update with allowlists: %v", err)
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

func TestAIAPIKeyUpdateAllModelsFalse(t *testing.T) {
	now := time.Now()
	ts := newAIAPIKeyUpdateServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	ts.keys = []v3.ListAIAPIKeysResponseEntry{{
		ID:             v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:           "alpha",
		Models:         &v3.AIAPIKeyModels{},
		Deployments:    &v3.AIAPIKeyDeployments{},
		AllModels:      boolPtr(true),
		AllDeployments: boolPtr(false),
		CreatedAT:      now,
		UpdatedAT:      now,
	}}

	var capturedRequest v3.UpdateAIAPIKeyRequest
	ts.captured = &capturedRequest

	c := &AIAPIKeyUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Key:                "alpha",
	}
	cmd := newAIAPIKeyUpdateCmd(t, c, "all-models")
	c.OutputFunc = func(output.Outputter, error) error { return nil }
	if err := c.CmdRun(cmd, nil); err != nil {
		t.Fatalf("api-key update --all-models=false: %v", err)
	}

	if capturedRequest.AllModels == nil || *capturedRequest.AllModels {
		t.Errorf("expected explicit all-models=false in request, got %v", capturedRequest.AllModels)
	}
	if capturedRequest.Models != nil {
		t.Errorf("expected models omitted from request, got %v", *capturedRequest.Models)
	}
	if capturedRequest.AllDeployments != nil {
		t.Errorf("expected all-deployments omitted from request when not passed, got %v", capturedRequest.AllDeployments)
	}
	if capturedRequest.Deployments != nil {
		t.Errorf("expected deployments omitted from request, got %v", *capturedRequest.Deployments)
	}
}

func TestAIAPIKeyUpdateAllDeploymentsFalse(t *testing.T) {
	now := time.Now()
	ts := newAIAPIKeyUpdateServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	ts.keys = []v3.ListAIAPIKeysResponseEntry{{
		ID:             v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:           "alpha",
		Models:         &v3.AIAPIKeyModels{},
		Deployments:    &v3.AIAPIKeyDeployments{},
		AllModels:      boolPtr(false),
		AllDeployments: boolPtr(true),
		CreatedAT:      now,
		UpdatedAT:      now,
	}}

	var capturedRequest v3.UpdateAIAPIKeyRequest
	ts.captured = &capturedRequest

	c := &AIAPIKeyUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Key:                "alpha",
	}
	cmd := newAIAPIKeyUpdateCmd(t, c, "all-deployments")
	c.OutputFunc = func(output.Outputter, error) error { return nil }
	if err := c.CmdRun(cmd, nil); err != nil {
		t.Fatalf("api-key update --all-deployments=false: %v", err)
	}

	if capturedRequest.AllDeployments == nil || *capturedRequest.AllDeployments {
		t.Errorf("expected explicit all-deployments=false in request, got %v", capturedRequest.AllDeployments)
	}
	if capturedRequest.Deployments != nil {
		t.Errorf("expected deployments omitted from request, got %v", *capturedRequest.Deployments)
	}
	if capturedRequest.AllModels != nil {
		t.Errorf("expected all-models omitted from request when not passed, got %v", capturedRequest.AllModels)
	}
	if capturedRequest.Models != nil {
		t.Errorf("expected models omitted from request, got %v", *capturedRequest.Models)
	}
}

func TestAIAPIKeyUpdateConflictingFlags(t *testing.T) {
	for _, tc := range []struct {
		c       *AIAPIKeyUpdateCmd
		changed []string
	}{
		{
			c:       &AIAPIKeyUpdateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Key: "alpha", AllModels: true, Models: []string{"llama-3-8b"}},
			changed: []string{"all-models"},
		},
		{
			c:       &AIAPIKeyUpdateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Key: "alpha", AllDeployments: true, Deployments: []string{"cccccccc-cccc-cccc-cccc-cccccccccccc"}},
			changed: []string{"all-deployments"},
		},
	} {
		cmd := newAIAPIKeyUpdateCmd(t, tc.c, tc.changed...)
		if err := tc.c.CmdRun(cmd, nil); err == nil {
			t.Errorf("expected conflict error for %+v", tc.c)
		}
	}
}

func TestAIAPIKeyUpdateOmitsUnsetAccessLists(t *testing.T) {
	now := time.Now()
	ts := newAIAPIKeyUpdateServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	ts.keys = []v3.ListAIAPIKeysResponseEntry{{
		ID:             v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:           "alpha",
		Models:         &v3.AIAPIKeyModels{},
		Deployments:    &v3.AIAPIKeyDeployments{},
		AllModels:      boolPtr(false),
		AllDeployments: boolPtr(false),
		CreatedAT:      now,
		UpdatedAT:      now,
	}}

	var capturedRequest v3.UpdateAIAPIKeyRequest
	ts.captured = &capturedRequest

	c := &AIAPIKeyUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Key:                "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		Models:             []string{"llama-3-8b"},
	}
	c.OutputFunc = func(output.Outputter, error) error { return nil }
	cmd := newAIAPIKeyUpdateCmd(t, c)
	if err := c.CmdRun(cmd, nil); err != nil {
		t.Fatalf("api-key update: %v", err)
	}

	if capturedRequest.Models == nil || len(*capturedRequest.Models) != 1 || (*capturedRequest.Models)[0] != "llama-3-8b" {
		t.Errorf("expected request models [llama-3-8b], got %v", capturedRequest.Models)
	}
	if capturedRequest.Deployments != nil {
		t.Errorf("expected deployments omitted from request, got %v", *capturedRequest.Deployments)
	}
	if capturedRequest.AllModels != nil || capturedRequest.AllDeployments != nil {
		t.Errorf("expected all-* flags omitted from request, got %v/%v", capturedRequest.AllModels, capturedRequest.AllDeployments)
	}
}

func TestAIAPIKeyUpdateWithoutProperties(t *testing.T) {
	c := &AIAPIKeyUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Key:                "alpha",
	}
	cmd := newAIAPIKeyUpdateCmd(t, c)
	if err := c.CmdRun(cmd, nil); err == nil {
		t.Fatal("expected error when no access list is updated")
	}
}

func TestAIAPIKeyUpdateUnknownKey(t *testing.T) {
	now := time.Now()
	ts := newAIAPIKeyUpdateServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	ts.keys = []v3.ListAIAPIKeysResponseEntry{{
		ID:        v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:      "alpha",
		CreatedAT: now,
		UpdatedAT: now,
	}}

	c := &AIAPIKeyUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Key:                "missing",
		Models:             []string{"llama-3-8b"},
	}
	cmd := newAIAPIKeyUpdateCmd(t, c)
	if err := c.CmdRun(cmd, nil); err == nil {
		t.Fatal("expected not found error")
	}
}
