// AI-modified by qwen3.8-27b-nvfp4-dflash2 - not reviewed yet
package dbaas

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type clickhouseTestServerOpts struct {
	ServerURL   string
	ServiceName string
	Service     *v3.DBAASServiceClickhouse
}

func setupClickhouseTestServer(t *testing.T, opts *clickhouseTestServerOpts) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		resp := v3.ListZonesResponse{Zones: []v3.Zone{{APIEndpoint: v3.Endpoint(opts.ServerURL), Name: v3.ZoneName("test-zone")}}}
		testutils.WriteJSON(t, w, http.StatusOK, resp)
	})

	mux.HandleFunc("/dbaas-service", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if opts.Service == nil {
			testutils.WriteJSON(t, w, http.StatusOK, v3.ListDBAASServicesResponse{})
			return
		}
		resp := v3.ListDBAASServicesResponse{
			DBAASServices: []v3.DBAASServiceCommon{
				{Name: v3.DBAASServiceName(opts.ServiceName), Type: "clickhouse"},
			},
		}
		testutils.WriteJSON(t, w, http.StatusOK, resp)
	})

	mux.HandleFunc("/dbaas-clickhouse/"+opts.ServiceName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			testutils.WriteJSON(t, w, http.StatusOK, *opts.Service)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/dbaas-clickhouse/"+opts.ServiceName+"/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			testutils.WriteJSON(t, w, http.StatusOK, v3.DBAASClickhouseUsers{
				Users: []v3.DBAASClickhouseUser{
					{Username: "avnadmin", Uuid: "22222222-2222-2222-2222-222222222222"},
					{Username: "testuser", Uuid: "33333333-3333-3333-3333-333333333333"},
				},
			})
		case http.MethodPost:
			var req v3.CreateDBAASClickhouseUserRequest
			body, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			require.NoError(t, json.Unmarshal(body, &req))
			testutils.WriteJSON(t, w, http.StatusOK, v3.DBAASUserClickhouseSecrets{
				Username: "testuser",
				Password: "generated-password",
			})
		}
	})

	mux.HandleFunc("/dbaas-clickhouse/"+opts.ServiceName+"/user/33333333-3333-3333-3333-333333333333", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-user-delete"), State: v3.OperationStateSuccess})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Operation wait (poll) endpoint
	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-user-delete"), State: v3.OperationStateSuccess})
	})

	ts := httptest.NewUnstartedServer(mux)
	ts.Start()
	// /zone reads opts.ServerURL at request time, so it is set after start.
	opts.ServerURL = ts.URL
	return ts
}

func TestDBAASClickhouseUserCreate(t *testing.T) {
	svc := &v3.DBAASServiceClickhouse{
		Users: []v3.DBAASClickhouseUser{
			{Username: "avnadmin", Uuid: "22222222-2222-2222-2222-222222222222"},
		},
	}
	opts := &clickhouseTestServerOpts{ServiceName: "testdb", Service: svc}
	ts := setupClickhouseTestServer(t, opts)
	defer ts.Close()

	testutils.SetupV3Client(t, ts.URL)

	rootCmd := &cobra.Command{}
	userCmd := &cobra.Command{Use: "user"}
	rootCmd.AddCommand(userCmd)
	c := &dbaasUserCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	err := exocmd.RegisterCLICommand(userCmd, c)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"user", "create", "testdb", "testuser", "--zone", "test-zone"})
	err = rootCmd.Execute()
	require.NoError(t, err)
}

func TestDBAASClickhouseUserDelete(t *testing.T) {
	svc := &v3.DBAASServiceClickhouse{}
	opts := &clickhouseTestServerOpts{ServiceName: "testdb", Service: svc}
	ts := setupClickhouseTestServer(t, opts)
	defer ts.Close()

	testutils.SetupV3Client(t, ts.URL)

	rootCmd := &cobra.Command{}
	userCmd := &cobra.Command{Use: "user"}
	rootCmd.AddCommand(userCmd)
	c := &dbaasUserDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	err := exocmd.RegisterCLICommand(userCmd, c)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"user", "delete", "testdb", "testuser", "--zone", "test-zone", "--force"})
	err = rootCmd.Execute()
	require.NoError(t, err)
}
