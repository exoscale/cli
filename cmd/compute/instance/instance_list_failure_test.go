package instance

import (
	"net/http"
	"net/http/httptest"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestInstanceListPartialFailure(t *testing.T) {
	instanceTypeID := v3.UUID("00000000-0000-0000-0000-000000000001")
	healthyZone := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/instance" {
			t.Fatalf("unexpected zone request path %q", r.URL.Path)
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListInstancesResponse{Instances: []v3.ListInstancesResponseInstances{{
			ID:           "00000000-0000-0000-0000-000000000002",
			Name:         "healthy-instance",
			InstanceType: &v3.InstanceType{ID: instanceTypeID},
			State:        v3.InstanceStateRunning,
		}}})
	}))
	defer healthyZone.Close()

	failingZone := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusBadRequest)
	}))
	defer failingZone.Close()

	control := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/zone":
			testutils.WriteJSON(t, w, http.StatusOK, v3.ListZonesResponse{Zones: []v3.Zone{
				{Name: "healthy-zone", APIEndpoint: v3.Endpoint(healthyZone.URL)},
				{Name: "failing-zone", APIEndpoint: v3.Endpoint(failingZone.URL)},
			}})
		case "/instance-type/" + string(instanceTypeID):
			testutils.WriteJSON(t, w, http.StatusOK, v3.InstanceType{
				ID: instanceTypeID, Family: "standard", Size: "medium",
			})
		default:
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer control.Close()
	testutils.SetupV3Client(t, control.URL)

	var got instanceListOutput
	cmd := &instanceListCmd{CliCommandSettings: exocmd.CliCommandSettings{
		OutputFunc: func(o output.Outputter, err error) error {
			if err != nil {
				t.Fatalf("output called with error: %v", err)
			}
			got = *o.(*instanceListOutput)
			return nil
		},
	}}

	if err := cmd.CmdRun(nil, nil); err == nil {
		t.Fatal("expected zone failure")
	}
	if len(got) != 1 || got[0].Name != "healthy-instance" {
		t.Fatalf("expected the healthy instance in partial output, got %+v", got)
	}
}
