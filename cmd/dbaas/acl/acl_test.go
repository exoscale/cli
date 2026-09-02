package acl

import (
	"net/http"
	"net/http/httptest"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type aclTestServerOpts struct {
	ServerURL   string
	ServiceName string
}

func setupAclTestServer(t *testing.T, opts *aclTestServerOpts) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		resp := v3.ListZonesResponse{Zones: []v3.Zone{{APIEndpoint: v3.Endpoint(opts.ServerURL), Name: v3.ZoneName("test-zone")}}}
		testutils.WriteJSON(t, w, http.StatusOK, resp)
	})

	mux.HandleFunc("/dbaas-clickhouse/"+opts.ServiceName+"/acl-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.DBAASClickhouseAclConfig{
			Users: []v3.DBAASClickhouseUserAclConfig{
				{
					Username: "avnadmin",
					Roles: []v3.DBAASClickhouseUserRole{
						{Name: "admin"},
					},
				},
			},
		})
	})

	ts := httptest.NewUnstartedServer(mux)
	ts.Start()
	// /zone reads opts.ServerURL at request time, so it is set after start.
	opts.ServerURL = ts.URL
	return ts
}

func TestDBAASClickhouseAclShow(t *testing.T) {
	opts := &aclTestServerOpts{ServiceName: "testdb"}
	ts := setupAclTestServer(t, opts)
	defer ts.Close()

	testutils.SetupV3Client(t, ts.URL)

	rootCmd := &cobra.Command{}
	aclCmd := &cobra.Command{Use: "acl"}
	rootCmd.AddCommand(aclCmd)
	c := &dbaasAclShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	err := exocmd.RegisterCLICommand(aclCmd, c)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"acl", "show", "testdb", "--zone", "test-zone"})
	err = rootCmd.Execute()
	require.NoError(t, err)
}
