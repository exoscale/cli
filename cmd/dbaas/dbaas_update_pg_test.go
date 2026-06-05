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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBAASPGUpdate(t *testing.T) {
	var gotReq v3.UpdateDBAASServicePGRequest
	ts := setupPGUpdateTestServer(t, &gotReq)
	defer ts.Close()

	testutils.SetupV3Client(t, ts.URL)

	c := &dbaasServiceUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}
	rootCmd := &cobra.Command{}
	err := exocmd.RegisterCLICommand(rootCmd, c)
	require.NoError(t, err)
	rootCmd.SetArgs([]string{
		"update", "testdb",
		"--zone", "test-zone",
		"--pg-migration-host", "123.123.123.123",
		"--pg-migration-port", "5432",
		"--pg-migration-username", "testsuperuser",
		"--pg-migration-password", "testpassword",
		"--pg-migration-method", "replication",
		"--pg-migration-ssl",
		"--pg-migration-dbname", "testdb",
		"--pg-migration-ignore-dbs", "telegraf,nagiosdb",
		"--force",
	})
	err = rootCmd.Execute()
	require.NoError(t, err)

	sslBool := true
	expReq := v3.UpdateDBAASServicePGRequest{
		Migration: &v3.UpdateDBAASServicePGRequestMigration{
			Dbname:    "testdb",
			Host:      "123.123.123.123",
			Port:      5432,
			Password:  "testpassword",
			Username:  "testsuperuser",
			Method:    "replication",
			SSL:       &sslBool,
			IgnoreDbs: "telegraf,nagiosdb",
		},
	}

	assert.Equal(t, expReq, gotReq)
}

func setupPGUpdateTestServer(t *testing.T, gotReq *v3.UpdateDBAASServicePGRequest) *httptest.Server {
	t.Helper()

	var ts *httptest.Server
	mux := http.NewServeMux()

	mux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp := v3.ListZonesResponse{Zones: []v3.Zone{{APIEndpoint: v3.Endpoint(ts.URL), Name: v3.ZoneName("test-zone")}}}
		testutils.WriteJSON(t, w, http.StatusOK, resp)
	})

	mux.HandleFunc("/dbaas-service", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := v3.ListDBAASServicesResponse{
				DBAASServices: []v3.DBAASServiceCommon{
					{
						Name: "testdb",
						Type: "pg",
					},
				},
			}
			testutils.WriteJSON(t, w, http.StatusOK, resp)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/dbaas-settings-pg", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := v3.GetDBAASSettingsPGResponse{}
			testutils.WriteJSON(t, w, http.StatusOK, resp)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/dbaas-postgres/testdb", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			err := r.Body.Close()
			require.NoError(t, err)
			err = json.Unmarshal(body, gotReq)
			require.NoError(t, err)
			testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-update"), State: v3.OperationStateSuccess})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-update"), State: v3.OperationStateSuccess})
	})

	ts = httptest.NewUnstartedServer(mux)
	ts.Start()
	return ts
}
