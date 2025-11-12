package dedicated_inference

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

type testServer struct {
	server *httptest.Server

	// simple in-memory fixtures
	models       []v3.ListModelsResponseEntry
	deployments  []v3.ListDeploymentsResponseEntry
	opPollCount  atomic.Int32
}

func newTestServer(t *testing.T) *testServer {
	ts := &testServer{}

	mux := http.NewServeMux()

 // Models endpoints
	mux.HandleFunc("/ai/model", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := v3.ListModelsResponse{Models: ts.models}
			writeJSON(t, w, http.StatusOK, resp)
		case http.MethodPost:
			// return operation pending
			op := v3.Operation{ID: v3.UUID("op-model-create"), State: v3.OperationStatePending}
			writeJSON(t, w, http.StatusOK, op)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	// GET/DELETE /ai/model/{id}
	mux.HandleFunc("/ai/model/", func(w http.ResponseWriter, r *http.Request) {
		id := path.Base(r.URL.Path)
		if r.Method == http.MethodGet {
			// find in fixtures
			for _, m := range ts.models {
				if string(m.ID) == id {
					resp := v3.GetModelResponse{
						ID:        m.ID,
						Name:      m.Name,
						Status:    v3.GetModelResponseStatus(m.Status),
						ModelSize: m.ModelSize,
						CreatedAT: m.CreatedAT,
						UpdatedAT: m.UpdatedAT,
					}
					writeJSON(t, w, http.StatusOK, resp)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodDelete {
			// Accept any UUID-like tail
			op := v3.Operation{ID: v3.UUID("op-model-delete"), State: v3.OperationStatePending}
			writeJSON(t, w, http.StatusOK, op)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Deployments endpoints
	mux.HandleFunc("/ai/deployment", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := v3.ListDeploymentsResponse{Deployments: ts.deployments}
			writeJSON(t, w, http.StatusOK, resp)
		case http.MethodPost:
			op := v3.Operation{ID: v3.UUID("op-deploy-create"), State: v3.OperationStatePending}
			writeJSON(t, w, http.StatusOK, op)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
 mux.HandleFunc("/ai/deployment/", func(w http.ResponseWriter, r *http.Request) {
		// patterns:
		// GET    /ai/deployment/{id}
		// DELETE /ai/deployment/{id}
		// POST   /ai/deployment/{id}/scale
		// GET    /ai/deployment/{id}/api-key
		// GET    /ai/deployment/{id}/logs
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
					resp := v3.GetDeploymentResponse{
						ID:            d.ID,
						Name:          d.Name,
						Status:        v3.GetDeploymentResponseStatus(d.Status),
						GpuType:       d.GpuType,
						GpuCount:      d.GpuCount,
						Replicas:      d.Replicas,
						ServiceLevel:  d.ServiceLevel,
						DeploymentURL: d.DeploymentURL,
						Model:         d.Model,
						CreatedAT:     d.CreatedAT,
						UpdatedAT:     d.UpdatedAT,
					}
					writeJSON(t, w, http.StatusOK, resp)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodDelete {
			op := v3.Operation{ID: v3.UUID("op-deploy-delete"), State: v3.OperationStatePending}
			writeJSON(t, w, http.StatusOK, op)
			return
		}
		if len(parts) == 2 && parts[1] == "scale" && r.Method == http.MethodPost {
			op := v3.Operation{ID: v3.UUID("op-deploy-scale"), State: v3.OperationStatePending}
			writeJSON(t, w, http.StatusOK, op)
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

	// Operation polling: pending -> success after a couple of polls
	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		// /operation/{id}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		count := ts.opPollCount.Add(1)
		state := v3.OperationStatePending
		if count >= 2 { // ensure at least one pending state is observed
			state = v3.OperationStateSuccess
		}
		op := v3.Operation{ID: v3.UUID(path.Base(r.URL.Path)), State: state}
		writeJSON(t, w, http.StatusOK, op)
	})

	// Generic 200 for zones endpoint used in other areas (not used here but may be called by SDK)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Unknown endpoints default 404 to expose unexpected calls
		w.WriteHeader(http.StatusNotFound)
	})

	ts.server = httptest.NewServer(mux)
	return ts
}

func writeJSON(t *testing.T, w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

func newTestClient(t *testing.T, endpoint string) *v3.Client {
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client = client.WithEndpoint(v3.Endpoint(endpoint))
	return client
}

func setup(t *testing.T, ts *testServer) func() {
	exocmd.GContext = context.Background()
	globalstate.Quiet = true // default quiet in tests; specific tests can toggle
	client := newTestClient(t, ts.server.URL)
	globalstate.EgoscaleV3Client = client
	return func() {
		ts.server.Close()
	}
}

// -------- Model tests --------

func TestModelShow(t *testing.T) {
	ts := newTestServer(t)
	defer setup(t, ts)()
	now := time.Now()
	ts.models = []v3.ListModelsResponseEntry{
		{ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1", Status: v3.ListModelsResponseEntryStatusReady, ModelSize: 123, CreatedAT: now, UpdatedAT: now},
	}
	// run show
	cmd := &modelShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), ID: "11111111-1111-1111-1111-111111111111"}
	var got modelShowOutput
	cmd.OutputFunc = func(o output.Outputter, err error) error {
		if err != nil { return err }
		got = *(o.(*modelShowOutput))
		return nil
	}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model show: %v", err)
	}
	if string(got.ID) != "11111111-1111-1111-1111-111111111111" || got.Name != "m1" || got.Status != v3.GetModelResponseStatusReady {
		t.Fatalf("unexpected model show output: %+v", got)
	}
}

func TestModelList(t *testing.T) {
	ts := newTestServer(t)
	defer setup(t, ts)()
	now := time.Now()
	ts.models = []v3.ListModelsResponseEntry{
		{ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1", Status: v3.ListModelsResponseEntryStatusReady, ModelSize: 0, CreatedAT: now, UpdatedAT: now},
		{ID: v3.UUID("22222222-2222-2222-2222-222222222222"), Name: "m2", Status: v3.ListModelsResponseEntryStatusCreating, ModelSize: 1234, CreatedAT: now, UpdatedAT: now},
	}
	cmd := &modelListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	// capture output through OutputFunc to ensure ToTable/Text/JSON paths are callable; here just run
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model list: %v", err)
	}
}

func TestModelCreateSuccessAndMissingName(t *testing.T) {
	ts := newTestServer(t)
	defer setup(t, ts)()
	// missing name
	cmd := &modelCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	if err := cmd.CmdRun(nil, nil); err == nil || !strings.Contains(err.Error(), "--name is required") {
		t.Fatalf("expected name required error, got %v", err)
	}
	// success
	globalstate.Quiet = true
	cmd = &modelCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Name: "foo", HuggingfaceToken: "tok"}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model create: %v", err)
	}
}

func TestModelDeleteInvalidUUIDAndSuccess(t *testing.T) {
	ts := newTestServer(t)
	defer setup(t, ts)()
	// invalid UUID
	cmd := &modelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), ID: "not-a-uuid"}
	if err := cmd.CmdRun(nil, nil); err == nil || !regexp.MustCompile(`invalid model ID`).MatchString(err.Error()) {
		t.Fatalf("expected invalid uuid error, got %v", err)
	}
	// success
	cmd = &modelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), ID: "33333333-3333-3333-3333-333333333333"}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model delete: %v", err)
	}
}

// -------- Deployment tests --------

func TestDeploymentList(t *testing.T) {
	ts := newTestServer(t)
	defer setup(t, ts)()
	now := time.Now()
	ts.deployments = []v3.ListDeploymentsResponseEntry{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "d1", Status: v3.ListDeploymentsResponseEntryStatusReady, GpuType: "gpua5000", GpuCount: 1, Replicas: 2, ServiceLevel: "pro", DeploymentURL: "https://u", Model: &v3.ModelRef{ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1"}, CreatedAT: now, UpdatedAT: now},
		{ID: v3.UUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), Name: "d2", Status: v3.ListDeploymentsResponseEntryStatusCreating, GpuType: "gpua5000", GpuCount: 2, Replicas: 1, ServiceLevel: "pro", DeploymentURL: "", Model: nil, CreatedAT: now, UpdatedAT: now},
	}
	cmd := &deploymentListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment list: %v", err)
	}
}

func TestDeploymentCreateByModelIDAndNameValidation(t *testing.T) {
	ts := newTestServer(t)
	defer setup(t, ts)()
	// missing required gpu type/count
	cmd := &deploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	if err := cmd.CmdRun(nil, nil); err == nil || !strings.Contains(err.Error(), "--gpu-type and --gpu-count are required") {
		t.Fatalf("expected gpu flags error, got %v", err)
	}
	// missing model flags
	cmd = &deploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), GPUType: "gpua5000", GPUCount: 1}
	if err := cmd.CmdRun(nil, nil); err == nil || !strings.Contains(err.Error(), "--model-id or --model-name is required") {
		t.Fatalf("expected model flag error, got %v", err)
	}
	// invalid model id
	cmd = &deploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), GPUType: "gpua5000", GPUCount: 1, ModelID: "bad"}
	if err := cmd.CmdRun(nil, nil); err == nil || !strings.Contains(err.Error(), "invalid --model-id") {
		t.Fatalf("expected invalid model id error, got %v", err)
	}
	// success with model id
	cmd = &deploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Name: "dep1", GPUType: "gpua5000", GPUCount: 1, Replicas: 2, ModelID: "11111111-1111-1111-1111-111111111111"}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment create (id): %v", err)
	}
	// success with model name
	cmd = &deploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Name: "dep2", GPUType: "gpua5000", GPUCount: 1, Replicas: 1, ModelName: "m1"}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment create (name): %v", err)
	}
}

func TestResolveDeploymentIDByIDAndName(t *testing.T) {
	ts := newTestServer(t)
	defer setup(t, ts)()
	now := time.Now()
	ts.deployments = []v3.ListDeploymentsResponseEntry{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", CreatedAT: now, UpdatedAT: now},
	}
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext
	// by ID
	id, err := resolveDeploymentID(ctx, client, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	if err != nil || string(id) != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("resolve by id failed: %v %v", id, err)
	}
	// by name
	id, err = resolveDeploymentID(ctx, client, "alpha")
	if err != nil || string(id) != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("resolve by name failed: %v %v", id, err)
	}
	// not found
	_, err = resolveDeploymentID(ctx, client, "missing")
	if err == nil || !strings.Contains(err.Error(), "deployment \"missing\" not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestDeploymentShowByIDAndName(t *testing.T) {
	ts := newTestServer(t)
	defer setup(t, ts)()
	now := time.Now()
	ts.deployments = []v3.ListDeploymentsResponseEntry{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", Status: v3.ListDeploymentsResponseEntryStatusReady, GpuType: "gpua5000", GpuCount: 1, Replicas: 1, ServiceLevel: "pro", DeploymentURL: "https://u", Model: &v3.ModelRef{ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1"}, CreatedAT: now, UpdatedAT: now},
	}
	// by ID
	cmd := &deploymentShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	var got deploymentShowOutput
	cmd.OutputFunc = func(o output.Outputter, err error) error {
		if err != nil { return err }
		got = *(o.(*deploymentShowOutput))
		return nil
	}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment show by id: %v", err)
	}
	if got.Name != "alpha" || got.GPUType != "gpua5000" || got.Status != v3.GetDeploymentResponseStatusReady {
		t.Fatalf("unexpected show output: %+v", got)
	}
	// by name
	cmd = &deploymentShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "alpha"}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("deployment show by name: %v", err)
	}
}

func TestDeploymentDeleteScaleRevealLogs(t *testing.T) {
	ts := newTestServer(t)
	defer setup(t, ts)()
	// populate name for resolution
	now := time.Now()
	ts.deployments = []v3.ListDeploymentsResponseEntry{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", CreatedAT: now, UpdatedAT: now},
	}
	// delete by name
	del := &deploymentDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "alpha"}
	if err := del.CmdRun(nil, nil); err != nil {
		t.Fatalf("delete: %v", err)
	}
	// scale by id
	sc := &deploymentScaleCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", Size: 3}
	if err := sc.CmdRun(nil, nil); err != nil {
		t.Fatalf("scale: %v", err)
	}
	// reveal api key
	reveal := &deploymentRevealAPIKeyCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "alpha"}
	// set output func to capture returned struct
	var got string
	reveal.OutputFunc = func(o output.Outputter, err error) error {
		if err != nil { return err }
		out := o.(*deploymentRevealAPIKeyOutput)
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
	logs := &deploymentLogsCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Deployment: "alpha"}
	if err := logs.CmdRun(nil, nil); err != nil {
		t.Fatalf("logs: %v", err)
	}
}
