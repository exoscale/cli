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

func TestDBAASOpensearchUpdate(t *testing.T) {
	enabled := true
	nestedObjectsLimit := int64(5000)
	replicas := int64(2)

	testCases := []struct {
		name   string
		args   []string
		expReq v3.UpdateDBAASServiceOpensearchRequest
	}{
		{
			name: "dashboard enabled",
			args: []string{"--opensearch-dashboard-enabled=true"},
			expReq: v3.UpdateDBAASServiceOpensearchRequest{
				OpensearchDashboards: &v3.UpdateDBAASServiceOpensearchRequestOpensearchDashboards{
					Enabled: &enabled,
				},
			},
		},
		{
			name: "dashboard request timeout",
			args: []string{"--opensearch-dashboard-request-timeout", "10000"},
			expReq: v3.UpdateDBAASServiceOpensearchRequest{
				OpensearchDashboards: &v3.UpdateDBAASServiceOpensearchRequestOpensearchDashboards{
					OpensearchRequestTimeout: 10000,
				},
			},
		},
		{
			name: "dashboard max old space size",
			args: []string{"--opensearch-dashboard-max-old-space-size", "256"},
			expReq: v3.UpdateDBAASServiceOpensearchRequest{
				OpensearchDashboards: &v3.UpdateDBAASServiceOpensearchRequestOpensearchDashboards{
					MaxOldSpaceSize: 256,
				},
			},
		},
		{
			name: "index template mapping nested objects limit",
			args: []string{"--opensearch-index-template-mapping-nested-objects-limit", "5000"},
			expReq: v3.UpdateDBAASServiceOpensearchRequest{
				IndexTemplate: &v3.UpdateDBAASServiceOpensearchRequestIndexTemplate{
					MappingNestedObjectsLimit: &nestedObjectsLimit,
				},
			},
		},
		{
			name: "index template number of replicas",
			args: []string{"--opensearch-index-template-number-of-replicas", "2"},
			expReq: v3.UpdateDBAASServiceOpensearchRequest{
				IndexTemplate: &v3.UpdateDBAASServiceOpensearchRequestIndexTemplate{
					NumberOfReplicas: &replicas,
				},
			},
		},
		{
			name: "index template number of shards",
			args: []string{"--opensearch-index-template-number-of-shards", "3"},
			expReq: v3.UpdateDBAASServiceOpensearchRequest{
				IndexTemplate: &v3.UpdateDBAASServiceOpensearchRequestIndexTemplate{
					NumberOfShards: 3,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gotReq v3.UpdateDBAASServiceOpensearchRequest
			ts := setupOpensearchTestServer(t, http.MethodPut, &gotReq)
			defer ts.Close()

			testutils.SetupV3Client(t, ts.URL)

			c := &dbaasServiceUpdateCmd{
				CliCommandSettings: exocmd.DefaultCLICmdSettings(),
			}
			rootCmd := &cobra.Command{}
			err := exocmd.RegisterCLICommand(rootCmd, c)
			require.NoError(t, err)
			rootCmd.SetArgs(append([]string{
				"update", "testdb",
				"--zone", "test-zone",
				"--force",
			}, tc.args...))
			err = rootCmd.Execute()
			require.NoError(t, err)

			assert.Equal(t, tc.expReq, gotReq)
		})
	}
}

func TestDBAASOpensearchCreate(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expDashboards v3.CreateDBAASServiceOpensearchRequestOpensearchDashboards
	}{
		{
			name: "dashboard request timeout",
			args: []string{"--opensearch-dashboard-request-timeout", "10000"},
			expDashboards: v3.CreateDBAASServiceOpensearchRequestOpensearchDashboards{
				OpensearchRequestTimeout: 10000,
			},
		},
		{
			name: "dashboard max old space size",
			args: []string{"--opensearch-dashboard-max-old-space-size", "256"},
			expDashboards: v3.CreateDBAASServiceOpensearchRequestOpensearchDashboards{
				MaxOldSpaceSize: 256,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gotReq v3.CreateDBAASServiceOpensearchRequest
			ts := setupOpensearchTestServer(t, http.MethodPost, &gotReq)
			defer ts.Close()

			testutils.SetupV3Client(t, ts.URL)

			c := &dbaasServiceCreateCmd{
				CliCommandSettings: exocmd.DefaultCLICmdSettings(),
			}
			rootCmd := &cobra.Command{}
			err := exocmd.RegisterCLICommand(rootCmd, c)
			require.NoError(t, err)
			rootCmd.SetArgs(append([]string{
				"create", "opensearch", "startup-4", "testdb",
				"--zone", "test-zone",
			}, tc.args...))
			err = rootCmd.Execute()
			require.NoError(t, err)

			require.NotNil(t, gotReq.OpensearchDashboards)
			assert.Equal(t, tc.expDashboards, *gotReq.OpensearchDashboards)
		})
	}
}

func setupOpensearchTestServer(t *testing.T, method string, gotReq any) *httptest.Server {
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
						Type: "opensearch",
					},
				},
			}
			testutils.WriteJSON(t, w, http.StatusOK, resp)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/dbaas-opensearch/testdb", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			body, _ := io.ReadAll(r.Body)
			err := r.Body.Close()
			require.NoError(t, err)
			err = json.Unmarshal(body, gotReq)
			require.NoError(t, err)
			testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-opensearch"), State: v3.OperationStateSuccess})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-opensearch"), State: v3.OperationStateSuccess})
	})

	ts = httptest.NewUnstartedServer(mux)
	ts.Start()
	return ts
}
