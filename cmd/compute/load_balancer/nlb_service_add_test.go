package load_balancer

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	v3 "github.com/exoscale/egoscale/v3"
)

const testServiceAddPath = "/load-balancer/" + testNLBID + "/service"

// testNLBServiceAddCmd returns an add command pre-filled with the same
// defaults that init() sets in production, which CmdRun does not apply.
func testNLBServiceAddCmd() *nlbServiceAddCmd {
	return &nlbServiceAddCmd{
		CliCommandSettings:  exocmd.DefaultCLICmdSettings(),
		NetworkLoadBalancer: "nlb-1",
		Name:                "web",
		Port:                443,
		Protocol:            "tcp",
		Strategy:            "round-robin",
		HealthcheckMode:     "tcp",
		HealthcheckInterval: 10,
		HealthcheckTimeout:  5,
		HealthcheckRetries:  1,
	}
}

func TestNLBServiceAddSelectorOnly(t *testing.T) {
	api := newNLBTestAPI(t)
	api.serveNLB()
	api.serveOperation(http.MethodPost, testServiceAddPath)

	c := testNLBServiceAddCmd()
	c.LabelSelector = map[string]string{"role": "web", "exoscale.com/managed-by": "karpenter"}

	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("service add: %v", err)
	}

	// a selector-only service has no Instance Pool, so there must be no pool
	// lookup, and the request must omit the instance-pool key entirely
	if n := api.called(http.MethodGet, "/instance-pool"); n > 0 {
		t.Errorf("expected no Instance Pool lookup when --instance-pool is not set, got %d requests", n)
	}
	body := api.body(http.MethodPost, testServiceAddPath)
	if body == nil {
		t.Fatal("expected a service add request to be issued")
	}
	if _, ok := body["instance-pool"]; ok {
		t.Errorf("instance-pool unexpectedly in request body: %v", body)
	}
	want := map[string]any{"role": "web", "exoscale.com/managed-by": "karpenter"}
	if got := bodySelectorLabels(t, body); !reflect.DeepEqual(got, want) {
		t.Errorf("labels: got %v, want %v", got, want)
	}
}

func TestNLBServiceAddPoolAndSelector(t *testing.T) {
	api := newNLBTestAPI(t)
	api.serveNLB()
	api.serveOperation(http.MethodPost, testServiceAddPath)
	api.Serve(http.MethodGet, "/instance-pool", v3.ListInstancePoolsResponse{
		InstancePools: []v3.InstancePool{{ID: v3.UUID(testInstancePoolID), Name: testInstancePoolName}},
	})

	c := testNLBServiceAddCmd()
	c.InstancePool = testInstancePoolName
	c.LabelSelector = map[string]string{"role": "web"}

	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("service add: %v", err)
	}

	body := api.body(http.MethodPost, testServiceAddPath)
	if body == nil {
		t.Fatal("expected a service add request to be issued")
	}
	pool, ok := body["instance-pool"].(map[string]any)
	if !ok {
		t.Fatalf("instance-pool missing from request body: %v", body)
	}
	if got := pool["id"]; got != testInstancePoolID {
		t.Errorf("instance-pool.id: got %v, want %q", got, testInstancePoolID)
	}
	want := map[string]any{"role": "web"}
	if got := bodySelectorLabels(t, body); !reflect.DeepEqual(got, want) {
		t.Errorf("labels: got %v, want %v", got, want)
	}
}

func TestNLBServiceAddRequiresBackendSource(t *testing.T) {
	api := newNLBTestAPI(t)

	c := testNLBServiceAddCmd()

	err := c.CmdRun(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "either --instance-pool or --label-selector is required") {
		t.Fatalf("expected backend source requirement error, got: %v", err)
	}
	if n := api.called(http.MethodPost, testServiceAddPath); n > 0 {
		t.Errorf("expected no service add request to be issued, got %d requests", n)
	}
}

// TestNLBServiceAddEmptySelectorRejected checks that an empty selector is
// rejected on create. Unlike update, where --label-selector '[=]' removes
// the selector, empty labels on create are a 400.
func TestNLBServiceAddEmptySelectorRejected(t *testing.T) {
	api := newNLBTestAPI(t)

	c := testNLBServiceAddCmd()
	// the map the framework parses from `--label-selector '[=]'`
	c.LabelSelector = map[string]string{"": ""}

	err := c.CmdRun(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "at least one label") {
		t.Fatalf("expected empty selector error, got: %v", err)
	}
	if n := api.called(http.MethodPost, testServiceAddPath); n > 0 {
		t.Errorf("expected no service add request to be issued, got %d requests", n)
	}
}
