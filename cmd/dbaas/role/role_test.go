package role

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

type roleTestServerOpts struct {
	ServerURL   string
	ServiceName string
}

func setupRoleTestServer(t *testing.T, opts *roleTestServerOpts) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		resp := v3.ListZonesResponse{Zones: []v3.Zone{{APIEndpoint: v3.Endpoint(opts.ServerURL), Name: v3.ZoneName("test-zone")}}}
		testutils.WriteJSON(t, w, http.StatusOK, resp)
	})

	mux.HandleFunc("/dbaas-clickhouse/"+opts.ServiceName+"/role", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.DBAASClickhouseRoles{
			Roles: []v3.DBAASClickhouseRole{
				{Name: v3.DBAASUserUsername("default_role"), Uuid: "11111111-1111-1111-1111-111111111111"},
			},
		})
	})

	mux.HandleFunc("/dbaas-clickhouse/"+opts.ServiceName+"/role/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-role-delete"), State: v3.OperationStateSuccess})
	})

	// Operation wait (poll) endpoint
	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-role-delete"), State: v3.OperationStateSuccess})
	})

	ts := httptest.NewUnstartedServer(mux)
	ts.Start()
	// /zone reads opts.ServerURL at request time, so it is set after start.
	opts.ServerURL = ts.URL
	return ts
}

func TestDBAASClickhouseRoleList(t *testing.T) {
	opts := &roleTestServerOpts{ServiceName: "testdb"}
	ts := setupRoleTestServer(t, opts)
	defer ts.Close()

	testutils.SetupV3Client(t, ts.URL)

	rootCmd := &cobra.Command{}
	roleCmd := &cobra.Command{Use: "role"}
	rootCmd.AddCommand(roleCmd)
	c := &dbaasRoleListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	err := exocmd.RegisterCLICommand(roleCmd, c)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"role", "list", "testdb", "--zone", "test-zone"})
	err = rootCmd.Execute()
	require.NoError(t, err)
}

func TestDBAASClickhouseRoleDelete(t *testing.T) {
	opts := &roleTestServerOpts{ServiceName: "testdb"}
	ts := setupRoleTestServer(t, opts)
	defer ts.Close()

	testutils.SetupV3Client(t, ts.URL)

	rootCmd := &cobra.Command{}
	roleCmd := &cobra.Command{Use: "role"}
	rootCmd.AddCommand(roleCmd)
	c := &dbaasRoleDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	err := exocmd.RegisterCLICommand(roleCmd, c)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{
		"role", "delete", "testdb", "11111111-1111-1111-1111-111111111111",
		"--zone", "test-zone", "--force",
	})
	err = rootCmd.Execute()
	require.NoError(t, err)
}
