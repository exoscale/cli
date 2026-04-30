package apikey

import (
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

type aiAPIKeyShowServer struct {
	server *httptest.Server
	keys   []v3.ListAIAPIKeysResponseEntry
}

func newAIAPIKeyShowServer(t *testing.T) *aiAPIKeyShowServer {
	t.Helper()
	ts := &aiAPIKeyShowServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/api-key", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListAIAPIKeysResponse{AIAPIKeys: ts.keys})
	})
	mux.HandleFunc("/ai/api-key/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := path.Base(r.URL.Path)
		for _, k := range ts.keys {
			if string(k.ID) == id {
				resp := v3.GetAIAPIKeyResponse{
					ID:             k.ID,
					Name:           k.Name,
					Models:         k.Models,
					Deployments:    k.Deployments,
					AllModels:      k.AllModels,
					AllDeployments: k.AllDeployments,
					CreatedAT:      k.CreatedAT,
					UpdatedAT:      k.UpdatedAT,
				}
				testutils.WriteJSON(t, w, http.StatusOK, resp)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func TestAIAPIKeyShowByIDAndName(t *testing.T) {
	now := time.Now()
	ts := newAIAPIKeyShowServer(t)
	defer ts.server.Close()
	testutils.SetupV3Client(t, ts.server.URL)

	ts.keys = []v3.ListAIAPIKeysResponseEntry{{
		ID:             v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:           "alpha",
		Models:         &v3.AIAPIKeyModels{"llama-3-8b"},
		Deployments:    &v3.AIAPIKeyDeployments{{ID: v3.UUID("cccccccc-cccc-cccc-cccc-cccccccccccc")}},
		AllModels:      boolPtr(true),
		AllDeployments: boolPtr(false),
		CreatedAT:      now,
		UpdatedAT:      now,
	}}

	show := func(key string) *AIAPIKeyShowOutput {
		var got *AIAPIKeyShowOutput
		cmd := &AIAPIKeyShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Key: key}
		cmd.OutputFunc = func(o output.Outputter, err error) error {
			if err != nil {
				return err
			}
			got = o.(*AIAPIKeyShowOutput)
			return nil
		}
		if err := cmd.CmdRun(nil, nil); err != nil {
			t.Fatalf("api-key show %q: %v", key, err)
		}
		return got
	}

	// by ID
	got := show("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	if got.Name != "alpha" || got.ID != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("unexpected show output: %+v", got)
	}
	if got.Models == nil || got.Models[0] != "llama-3-8b" {
		t.Errorf("unexpected models: %v", got.Models)
	}
	if got.Deployments == nil || got.Deployments[0] != "cccccccc-cccc-cccc-cccc-cccccccccccc" {
		t.Errorf("unexpected deployments: %v", got.Deployments)
	}
	if !got.AllModels || got.AllDeployments {
		t.Errorf("unexpected all flags, got all_models=%v all_deployments=%v", got.AllModels, got.AllDeployments)
	}

	// by name
	if show("alpha").ID != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatal("show by name failed")
	}

	// not found
	cmd := &AIAPIKeyShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Key: "missing"}
	if err := cmd.CmdRun(nil, nil); err == nil {
		t.Fatal("expected not found error")
	}
}
