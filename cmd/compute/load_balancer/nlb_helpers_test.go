package load_balancer

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

const (
	testNLBID            = "8f3a0000-0000-4000-8000-0000000000a1"
	testNLBName          = "nlb-1"
	testInstancePoolID   = "8f3a0000-0000-4000-8000-0000000000b1"
	testInstancePoolName = "pool-1"
	testPoolServiceID    = "8f3a0000-0000-4000-8000-0000000000c1"
	testPoolServiceName  = "pool-svc"
	testSelectorSvcID    = "8f3a0000-0000-4000-8000-0000000000c2"
	testSelectorSvcName  = "selector-svc"
	testOperationID      = "00000000-0000-0000-0000-0000000000f1"
)

// nlbRequest is a single request made against the mock API. Body is the decoded
// JSON request body (nil for requests without one).
type nlbRequest struct {
	Method string
	Path   string
	Body   map[string]any
}

// nlbTestAPI is a mock NLB API that records every request, so tests can make
// assertions on calls. Responses are registered per method and path with Serve.
type nlbTestAPI struct {
	requests  []nlbRequest
	responses map[string]any
}

// Serve registers a JSON response for "<METHOD> <path>".
func (a *nlbTestAPI) Serve(method, path string, v any) {
	a.responses[method+" "+path] = v
}

// serveOperation marks the given request as a successful asynchronous call.
// The real API answers writes with a pending Operation that the client polls
// until it completes. The mock returns the operation already completed, so
// client.Wait returns without polling and no GET /operation request happens.
func (a *nlbTestAPI) serveOperation(method, path string) {
	a.Serve(method, path, v3.Operation{
		ID:    v3.UUID(testOperationID),
		State: v3.OperationStateSuccess,
	})
}

// serveNLB registers a single Network Load Balancer with the given services.
func (a *nlbTestAPI) serveNLB(services ...v3.LoadBalancerService) {
	a.Serve(http.MethodGet, "/load-balancer", v3.ListLoadBalancersResponse{
		LoadBalancers: []v3.LoadBalancer{{
			ID:       v3.UUID(testNLBID),
			Name:     testNLBName,
			Services: services,
		}},
	})
}

// baseService returns a service with valid default settings.
func baseService() v3.LoadBalancerService {
	return v3.LoadBalancerService{
		Healthcheck: &v3.LoadBalancerServiceHealthcheck{
			Mode:     v3.LoadBalancerServiceHealthcheckModeTCP,
			Port:     8443,
			Interval: 10,
			Timeout:  2,
			Retries:  1,
		},
		Protocol:   v3.LoadBalancerServiceProtocolTCP,
		Strategy:   v3.LoadBalancerServiceStrategyRoundRobin,
		Port:       443,
		TargetPort: 8443,
		State:      v3.LoadBalancerServiceStateRunning,
	}
}

// poolService returns a service backed by the test Instance Pool.
func poolService(name, id string) v3.LoadBalancerService {
	svc := baseService()
	svc.ID = v3.UUID(id)
	svc.Name = name
	svc.InstancePool = &v3.InstancePool{ID: v3.UUID(testInstancePoolID)}
	return svc
}

// selectorService returns a service backed only by a label selector.
func selectorService(name, id string, labels map[string]string) v3.LoadBalancerService {
	svc := baseService()
	svc.ID = v3.UUID(id)
	svc.Name = name
	svc.LabelSelector = &v3.LoadBalancerServiceLabelSelector{Labels: labels}
	return svc
}

// called returns the number of requests made to the given path with method.
func (a *nlbTestAPI) called(method, path string) int {
	var n int
	for _, r := range a.requests {
		if r.Method == method && r.Path == path {
			n++
		}
	}
	return n
}

// body returns the JSON body of the last matching request with method and path,
// or nil if the request was never made.
func (a *nlbTestAPI) body(method, path string) map[string]any {
	var body map[string]any
	for _, r := range a.requests {
		if r.Method == method && r.Path == path {
			body = r.Body
		}
	}
	return body
}

// newNLBTestAPI starts a mock NLB API server and points the global v3 client
// at it. The server is closed when the test ends. Tests register mock
// responses with Serve, serveNLB and serveOperation. Unregistered requests
// return 404.
func newNLBTestAPI(t *testing.T) *nlbTestAPI {
	t.Helper()

	api := &nlbTestAPI{responses: map[string]any{}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := nlbRequest{Method: r.Method, Path: r.URL.Path}
		if r.Body != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read request body: %v", err)
			} else if len(body) > 0 {
				if err := json.Unmarshal(body, &req.Body); err != nil {
					t.Errorf("unmarshal request body: %v", err)
				}
			}
		}
		api.requests = append(api.requests, req)

		resp, ok := api.responses[req.Method+" "+req.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, resp)
	}))
	t.Cleanup(srv.Close)

	testutils.SetupV3Client(t, srv.URL)
	return api
}

func nlbServicePath(serviceID string) string {
	return "/load-balancer/" + testNLBID + "/service/" + serviceID
}

// bodySelectorLabels returns the label-selector labels of a captured
// request body, or fails the test if they are missing.
func bodySelectorLabels(t *testing.T, body map[string]any) map[string]any {
	t.Helper()

	selector, ok := body["label-selector"].(map[string]any)
	if !ok {
		t.Fatalf("label-selector missing from request body: %v", body)
	}
	labels, ok := selector["labels"].(map[string]any)
	if !ok {
		t.Fatalf("label-selector.labels missing from request body: %v", selector)
	}
	return labels
}

// newServiceUpdateFlagSet builds a cobra command carrying the update
// command's generated flags, so cmd.Flags().Changed() behaves as it does in
// production.
func newServiceUpdateFlagSet(t *testing.T, c *nlbServiceUpdateCmd, changed ...string) *cobra.Command {
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
