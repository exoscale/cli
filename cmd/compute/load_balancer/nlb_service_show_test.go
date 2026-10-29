package load_balancer

import (
	"reflect"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
)

// TestNLBServiceShowSelectorOnlyService checks that show handles services
// returned without an instance-pool key (the new shape for label-selector
// only services). The command must not fail on the missing Instance Pool,
// and must display the selector.
func TestNLBServiceShowSelectorOnlyService(t *testing.T) {
	api := newNLBTestAPI(t)
	api.serveNLB(
		poolService(testPoolServiceName, testPoolServiceID),
		selectorService(testSelectorSvcName, testSelectorSvcID, map[string]string{"role": "web"}),
	)

	// selector-only service, with no Instance Pool and a label selector
	var got *nlbServiceShowOutput
	cmd := &nlbServiceShowCmd{
		CliCommandSettings:  exocmd.DefaultCLICmdSettings(),
		NetworkLoadBalancer: testNLBName,
		Service:             testSelectorSvcName,
	}
	// keep the produced output instead of printing it
	cmd.OutputFunc = func(o output.Outputter, err error) error {
		got = o.(*nlbServiceShowOutput)
		return err
	}

	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("service show: %v", err)
	}
	if got.InstancePoolID != "" {
		t.Errorf("instance_pool_id: got %q, want empty", got.InstancePoolID)
	}
	wantLabels := map[string]string{"role": "web"}
	if !reflect.DeepEqual(got.LabelSelector, wantLabels) {
		t.Errorf("label_selector: got %v, want %v", got.LabelSelector, wantLabels)
	}

	// pool-backed service, with an Instance Pool and no label selector
	got = nil
	cmd = &nlbServiceShowCmd{
		CliCommandSettings:  exocmd.DefaultCLICmdSettings(),
		NetworkLoadBalancer: testNLBName,
		Service:             testPoolServiceName,
	}
	cmd.OutputFunc = func(o output.Outputter, err error) error {
		got = o.(*nlbServiceShowOutput)
		return err
	}

	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("service show: %v", err)
	}
	if got.InstancePoolID != testInstancePoolID {
		t.Errorf("instance_pool_id: got %q, want %q", got.InstancePoolID, testInstancePoolID)
	}
	if got.LabelSelector != nil {
		t.Errorf("label_selector: got %v, want nil", got.LabelSelector)
	}
}
