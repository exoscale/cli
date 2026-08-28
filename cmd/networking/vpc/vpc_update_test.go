package vpc

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

const testVPCID = "8f3a0000-0000-4000-8000-000000000001"

// newUpdateFlagSet builds a cobra command carrying the update command's
// generated flags, so cmd.Flags().Changed() behaves as it does in production.
func newUpdateFlagSet(t *testing.T, c *vpcUpdateCmd, changed ...string) *cobra.Command {
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

func vpcListHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListVpcsResponse{
			Vpcs: []v3.ListVpcEntry{{ID: v3.UUID(testVPCID), Name: "prod"}},
		})
	}
}

// TestVPCUpdateOmitsUnsetFields asserts that flags the user did not pass are
// left out of the PUT body entirely, rather than being sent as empty strings
// (which would wipe the description server-side).
func TestVPCUpdateOmitsUnsetFields(t *testing.T) {
	var rawBody []byte
	var called bool

	mux := http.NewServeMux()
	mux.HandleFunc("/vpc", vpcListHandler(t))
	mux.HandleFunc("/vpc/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		called = true
		rawBody, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
		testutils.WriteJSON(t, w, http.StatusOK, v3.Vpc{ID: v3.UUID(testVPCID), Name: "renamed"})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	testutils.SetupV3Client(t, srv.URL)

	c := &vpcUpdateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), VPC: "prod"}
	cmd := newUpdateFlagSet(t, c, "name")
	c.VPC = "prod"
	c.Name = "renamed"

	if err := c.CmdRun(cmd, nil); err != nil {
		t.Fatalf("vpc update: %v", err)
	}

	if !called {
		t.Fatal("expected an update request to be issued")
	}

	var body map[string]any
	if err := json.Unmarshal(rawBody, &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if got := body["name"]; got != "renamed" {
		t.Errorf("name: got %v, want %q", got, "renamed")
	}
	if _, ok := body["description"]; ok {
		t.Errorf("description must be omitted when --description is not set, body was %s", rawBody)
	}
}
