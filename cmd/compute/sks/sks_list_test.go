package sks

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestSksListZoneFilterSetsZone(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/zone":
			testutils.WriteJSON(t, w, http.StatusOK, v3.ListZonesResponse{Zones: []v3.Zone{
				{Name: "ch-dk-2", APIEndpoint: v3.Endpoint(srv.URL)},
			}})
		case "/sks-cluster":
			testutils.WriteJSON(t, w, http.StatusOK, v3.ListSKSClustersResponse{SKSClusters: []v3.SKSCluster{
				{ID: "00000000-0000-0000-0000-000000000001", Name: "cluster"},
			}})
		default:
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer srv.Close()
	testutils.SetupV3Client(t, srv.URL)
	globalstate.OutputFormat = "json"

	var stdout, stderr bytes.Buffer
	require.NoError(t, runSksList(&sksListCmd{Zone: "ch-dk-2"}, &stdout, &stderr))
	require.Contains(t, stdout.String(), `"zone":"ch-dk-2"`)
}
