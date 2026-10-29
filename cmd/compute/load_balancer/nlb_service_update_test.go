package load_balancer

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
)

func testNLBServiceUpdateCmd(service string) *nlbServiceUpdateCmd {
	return &nlbServiceUpdateCmd{
		CliCommandSettings:  exocmd.DefaultCLICmdSettings(),
		NetworkLoadBalancer: testNLBName,
		Service:             service,
	}
}

// TestNLBServiceUpdateReplacesSelector checks that the updated selector
// replaces the previous one, and that keys like exoscale.com/managed-by are
// sent as-is, keeping their prefix and case.
func TestNLBServiceUpdateReplacesSelector(t *testing.T) {
	api := newNLBTestAPI(t)
	api.serveNLB(poolService(testPoolServiceName, testPoolServiceID))
	api.serveOperation(http.MethodPut, nlbServicePath(testPoolServiceID))

	c := testNLBServiceUpdateCmd(testPoolServiceName)
	c.LabelSelector = map[string]string{"role": "api", "exoscale.com/managed-by": "karpenter"}
	cmd := newServiceUpdateFlagSet(t, c, "label-selector")

	if err := c.CmdRun(cmd, nil); err != nil {
		t.Fatalf("service update: %v", err)
	}

	path := nlbServicePath(testPoolServiceID)
	body := api.body(http.MethodPut, path)
	if body == nil {
		t.Fatal("expected an update request to be issued")
	}
	want := map[string]any{"role": "api", "exoscale.com/managed-by": "karpenter"}
	if got := bodySelectorLabels(t, body); !reflect.DeepEqual(got, want) {
		t.Errorf("labels: got %v, want %v", got, want)
	}
}

// TestNLBServiceUpdateLeavesSelectorUntouched checks that a flag the user did
// not pass is omitted from the PUT body entirely. Sending an explicit null
// would do nothing server-side, so the "leave untouched" contract relies on
// the omission.
func TestNLBServiceUpdateLeavesSelectorUntouched(t *testing.T) {
	api := newNLBTestAPI(t)
	api.serveNLB(poolService(testPoolServiceName, testPoolServiceID))
	api.serveOperation(http.MethodPut, nlbServicePath(testPoolServiceID))

	c := testNLBServiceUpdateCmd(testPoolServiceName)
	c.Description = "hello"
	cmd := newServiceUpdateFlagSet(t, c, "description")

	if err := c.CmdRun(cmd, nil); err != nil {
		t.Fatalf("service update: %v", err)
	}

	body := api.body(http.MethodPut, nlbServicePath(testPoolServiceID))
	if body == nil {
		t.Fatal("expected an update request to be issued")
	}
	if got := body["description"]; got != "hello" {
		t.Errorf("description: got %v, want %q", got, "hello")
	}
	if _, ok := body["label-selector"]; ok {
		t.Errorf("label-selector unexpectedly in request body: %v", body)
	}
}

// TestNLBServiceUpdateRemovesSelector checks that --label-selector '[=]'
// sends an explicitly empty labels object, which is the API's encoding for
// "remove the selector". A null or omitted key would not remove it.
func TestNLBServiceUpdateRemovesSelector(t *testing.T) {
	api := newNLBTestAPI(t)
	api.serveNLB(poolService(testPoolServiceName, testPoolServiceID))
	api.serveOperation(http.MethodPut, nlbServicePath(testPoolServiceID))

	c := testNLBServiceUpdateCmd(testPoolServiceName)
	// the map the framework parses from `--label-selector '[=]'`
	c.LabelSelector = map[string]string{"": ""}
	cmd := newServiceUpdateFlagSet(t, c, "label-selector")

	if err := c.CmdRun(cmd, nil); err != nil {
		t.Fatalf("service update: %v", err)
	}

	body := api.body(http.MethodPut, nlbServicePath(testPoolServiceID))
	if body == nil {
		t.Fatal("expected an update request to be issued")
	}
	selector, ok := body["label-selector"].(map[string]any)
	if !ok {
		t.Fatalf("label-selector missing from request body: %v", body)
	}
	labels, ok := selector["labels"].(map[string]any)
	if !ok {
		t.Fatalf("label-selector.labels must be an empty object, got: %v", selector["labels"])
	}
	if len(labels) != 0 {
		t.Errorf("label-selector.labels must be empty, got: %v", labels)
	}
}

// TestNLBServiceUpdateRemoveSelectorOnPoollessService checks that removing
// the selector from a service with no Instance Pool is rejected upfront,
// since it would leave the service with zero backends. Without the guard the
// API would answer with a 400.
func TestNLBServiceUpdateRemoveSelectorOnPoollessService(t *testing.T) {
	api := newNLBTestAPI(t)
	api.serveNLB(selectorService(testSelectorSvcName, testSelectorSvcID, map[string]string{"role": "web"}))

	c := testNLBServiceUpdateCmd(testSelectorSvcName)
	c.LabelSelector = map[string]string{"": ""}
	cmd := newServiceUpdateFlagSet(t, c, "label-selector")

	err := c.CmdRun(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "has no Instance Pool") {
		t.Fatalf("expected pool-less removal error, got: %v", err)
	}
	if n := api.called(http.MethodPut, nlbServicePath(testSelectorSvcID)); n > 0 {
		t.Errorf("expected no update request to be issued, got %d requests", n)
	}
}
