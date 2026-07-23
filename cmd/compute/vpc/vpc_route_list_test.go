package vpc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

const testSubnetID = "8f3a0000-0000-4000-8000-000000000002"

func routeListMux(t *testing.T, hit *string) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/vpc", vpcListHandler(t))

	// VPC-wide routes: GET /vpc/{id}/route
	mux.HandleFunc("/vpc/"+testVPCID+"/route", func(w http.ResponseWriter, _ *http.Request) {
		*hit = "vpc"
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListVpcRoutesResponse{
			Routes: []v3.ListRouteEntry{
				{ID: v3.UUID("00000000-0000-4000-8000-00000000000a"), Kind: v3.ListRouteEntryKindSubnet, Destination: "10.0.2.0/24"},
				{ID: v3.UUID("00000000-0000-4000-8000-00000000000b"), Kind: v3.ListRouteEntryKindVpc, Destination: "10.0.9.0/24"},
				{ID: v3.UUID("00000000-0000-4000-8000-00000000000c"), Kind: v3.ListRouteEntryKindSubnet, Destination: "10.0.1.0/24"},
				{ID: v3.UUID("00000000-0000-4000-8000-00000000000d"), Kind: v3.ListRouteEntryKindVpc, Destination: "10.0.8.0/24"},
			},
		})
	})

	// Subnet listing, for --subnet resolution
	mux.HandleFunc("/vpc/"+testVPCID+"/subnet", func(w http.ResponseWriter, _ *http.Request) {
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListSubnetsResponse{
			Subnets: []v3.ListSubnetEntry{{ID: v3.UUID(testSubnetID), Name: "web"}},
		})
	})

	// Subnet-scoped routes: GET /vpc/{id}/subnet/{sid}/route
	mux.HandleFunc("/vpc/"+testVPCID+"/subnet/"+testSubnetID+"/route", func(w http.ResponseWriter, _ *http.Request) {
		*hit = "subnet"
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListRoutesResponse{
			Routes: []v3.ListRouteEntry{
				{ID: v3.UUID("00000000-0000-4000-8000-00000000000e"), Kind: v3.ListRouteEntryKindSubnet, Destination: "10.0.1.0/24"},
			},
		})
	})

	return mux
}

// TestVPCRouteListEndpointSelection asserts --subnet switches the CLI between
// the VPC-wide and the Subnet-scoped route endpoints.
func TestVPCRouteListEndpointSelection(t *testing.T) {
	for _, tc := range []struct {
		name    string
		subnet  string
		wantHit string
	}{
		{name: "without --subnet", subnet: "", wantHit: "vpc"},
		{name: "with --subnet", subnet: "web", wantHit: "subnet"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var hit string
			srv := httptest.NewServer(routeListMux(t, &hit))
			defer srv.Close()
			testutils.SetupV3Client(t, srv.URL)

			c := &vpcRouteListCmd{
				CliCommandSettings: exocmd.DefaultCLICmdSettings(),
				VPC:                "prod",
				Subnet:             tc.subnet,
			}

			if _, err := c.list(); err != nil {
				t.Fatalf("route list: %v", err)
			}

			if hit != tc.wantHit {
				t.Errorf("endpoint: got %q, want %q", hit, tc.wantHit)
			}
		})
	}
}

// TestVPCRouteListOrdersVpcKindFirst asserts VPC routes are emitted before
// Subnet routes, each group ordered by destination.
func TestVPCRouteListOrdersVpcKindFirst(t *testing.T) {
	var hit string
	srv := httptest.NewServer(routeListMux(t, &hit))
	defer srv.Close()
	testutils.SetupV3Client(t, srv.URL)

	c := &vpcRouteListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), VPC: "prod"}

	out, err := c.list()
	if err != nil {
		t.Fatalf("route list: %v", err)
	}

	// Vpc-kind routes first, each group ordered by destination.
	want := []vpcRouteListItemOutput{
		{Kind: "Vpc", Destination: "10.0.8.0/24"},
		{Kind: "Vpc", Destination: "10.0.9.0/24"},
		{Kind: "Subnet", Destination: "10.0.1.0/24"},
		{Kind: "Subnet", Destination: "10.0.2.0/24"},
	}

	if len(*out) != len(want) {
		t.Fatalf("got %d routes, want %d: %+v", len(*out), len(want), *out)
	}

	for i, w := range want {
		got := (*out)[i]
		if got.Kind != w.Kind || got.Destination != w.Destination {
			t.Errorf("route %d: got %s/%s, want %s/%s",
				i, got.Kind, got.Destination, w.Kind, w.Destination)
		}
	}
}
