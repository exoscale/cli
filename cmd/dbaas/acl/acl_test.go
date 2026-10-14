package acl

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	ServiceType string
	// HTTP status of the service list, 200 if unset.
	ListServicesStatus int

	KafkaAcls     v3.DBAASKafkaAcls
	OpensearchAcl v3.DBAASOpensearchAclConfig

	// Requests received by the server, as "METHOD PATH".
	Requests []string
	// Last Kafka ACL entries sent to the server.
	KafkaTopicAclCreate          *v3.DBAASKafkaTopicAclEntry
	KafkaSchemaRegistryAclCreate *v3.DBAASKafkaSchemaRegistryAclEntry
}

func setupAclTestServer(t *testing.T, opts *aclTestServerOpts) *httptest.Server {
	t.Helper()

	if opts.ServiceType == "" {
		opts.ServiceType = "clickhouse"
	}

	mux := http.NewServeMux()

	writeOperation := func(w http.ResponseWriter) {
		testutils.WriteJSON(t, w, http.StatusOK, v3.Operation{ID: v3.UUID("op-acl"), State: v3.OperationStateSuccess})
	}

	mux.HandleFunc("/zone", func(w http.ResponseWriter, r *http.Request) {
		resp := v3.ListZonesResponse{Zones: []v3.Zone{{APIEndpoint: v3.Endpoint(opts.ServerURL), Name: v3.ZoneName("test-zone")}}}
		testutils.WriteJSON(t, w, http.StatusOK, resp)
	})

	mux.HandleFunc("/dbaas-service", func(w http.ResponseWriter, r *http.Request) {
		if opts.ListServicesStatus != 0 && opts.ListServicesStatus != http.StatusOK {
			testutils.WriteJSON(t, w, opts.ListServicesStatus, map[string]string{"message": "forbidden"})
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, v3.ListDBAASServicesResponse{
			DBAASServices: []v3.DBAASServiceCommon{{
				Name: v3.DBAASServiceName(opts.ServiceName),
				Type: v3.DBAASServiceTypeName(opts.ServiceType),
			}},
		})
	})

	mux.HandleFunc("/operation/", func(w http.ResponseWriter, r *http.Request) {
		writeOperation(w)
	})

	mux.HandleFunc("/dbaas-clickhouse/"+opts.ServiceName+"/acl-config", func(w http.ResponseWriter, r *http.Request) {
		opts.Requests = append(opts.Requests, r.Method+" "+r.URL.Path)
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

	mux.HandleFunc("/dbaas-kafka/"+opts.ServiceName+"/", func(w http.ResponseWriter, r *http.Request) {
		opts.Requests = append(opts.Requests, r.Method+" "+strings.TrimPrefix(r.URL.Path, "/dbaas-kafka/"+opts.ServiceName))

		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/acl-config"):
			testutils.WriteJSON(t, w, http.StatusOK, opts.KafkaAcls)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/topic/acl-config"):
			opts.KafkaTopicAclCreate = &v3.DBAASKafkaTopicAclEntry{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(opts.KafkaTopicAclCreate))
			writeOperation(w)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/schema-registry/acl-config"):
			opts.KafkaSchemaRegistryAclCreate = &v3.DBAASKafkaSchemaRegistryAclEntry{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(opts.KafkaSchemaRegistryAclCreate))
			writeOperation(w)
		case r.Method == http.MethodDelete:
			writeOperation(w)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/dbaas-opensearch/"+opts.ServiceName+"/acl-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		testutils.WriteJSON(t, w, http.StatusOK, opts.OpensearchAcl)
	})

	ts := httptest.NewUnstartedServer(mux)
	ts.Start()
	// /zone reads opts.ServerURL at request time, so it is set after start.
	opts.ServerURL = ts.URL
	return ts
}

// runAclCmd registers the "acl" subcommands on a fresh command tree and runs
// "acl ARGS... --zone test-zone".
func runAclCmd(t *testing.T, args ...string) error {
	t.Helper()

	rootCmd := &cobra.Command{SilenceUsage: true, SilenceErrors: true}
	aclCmd := &cobra.Command{Use: "acl"}
	rootCmd.AddCommand(aclCmd)

	for _, c := range []interface {
		CmdAliases() []string
		CmdShort() string
		CmdLong() string
		CmdPreRun(*cobra.Command, []string) error
		CmdRun(*cobra.Command, []string) error
	}{
		&dbaasAclShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()},
		&dbaasAclListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()},
		&dbaasAclCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()},
		&dbaasAclDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()},
	} {
		require.NoError(t, exocmd.RegisterCLICommand(aclCmd, c))
	}

	rootCmd.SetArgs(append(append([]string{"acl"}, args...), "--zone", "test-zone"))
	return rootCmd.Execute()
}

func testKafkaAcls() v3.DBAASKafkaAcls {
	return v3.DBAASKafkaAcls{
		TopicAcl: []v3.DBAASKafkaTopicAclEntry{
			{ID: "topic-1", Username: "alice", Topic: "orders", Permission: v3.DBAASKafkaTopicAclEntryPermissionRead},
			{ID: "topic-2", Username: "alice", Topic: "orders", Permission: v3.DBAASKafkaTopicAclEntryPermissionWrite},
			{ID: "topic-3", Username: "bob", Topic: "*", Permission: v3.DBAASKafkaTopicAclEntryPermissionAdmin},
		},
		SchemaRegistryAcl: []v3.DBAASKafkaSchemaRegistryAclEntry{
			{ID: "schema-1", Username: "alice", Resource: "Subject:orders", Permission: v3.DBAASKafkaSchemaRegistryAclEntryPermissionSchemaRegistryRead},
		},
	}
}

func testOpensearchAcl() v3.DBAASOpensearchAclConfig {
	enabled := true
	return v3.DBAASOpensearchAclConfig{
		AclEnabled: &enabled,
		Acls: []v3.DBAASOpensearchAclConfigAcls{
			{
				Username: "alice",
				Rules: []v3.DBAASOpensearchAclConfigAclsRules{
					{Index: "logs-*", Permission: v3.EnumOpensearchRulePermissionRead},
					{Index: "metrics-*", Permission: v3.EnumOpensearchRulePermissionWrite},
				},
			},
			{
				Username: "bob",
				Rules: []v3.DBAASOpensearchAclConfigAclsRules{
					{Index: "*", Permission: v3.EnumOpensearchRulePermissionAdmin},
				},
			},
		},
	}
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

// API keys that may read the ClickHouse ACL but not list the services keep
// working, as before Kafka and OpenSearch support.
func TestDBAASClickhouseAclShowWithoutServiceList(t *testing.T) {
	opts := &aclTestServerOpts{ServiceName: "testdb", ListServicesStatus: http.StatusForbidden}
	ts := setupAclTestServer(t, opts)
	defer ts.Close()

	testutils.SetupV3Client(t, ts.URL)

	require.NoError(t, runAclCmd(t, "show", "testdb"))
	require.Equal(t, []string{"GET /dbaas-clickhouse/testdb/acl-config"}, opts.Requests)
}

func TestDBAASAclShow(t *testing.T) {
	for _, dbType := range []string{"kafka", "opensearch"} {
		t.Run(dbType, func(t *testing.T) {
			opts := &aclTestServerOpts{
				ServiceName:   "testdb",
				ServiceType:   dbType,
				KafkaAcls:     testKafkaAcls(),
				OpensearchAcl: testOpensearchAcl(),
			}
			ts := setupAclTestServer(t, opts)
			defer ts.Close()

			testutils.SetupV3Client(t, ts.URL)

			require.NoError(t, runAclCmd(t, "show", "testdb"))
		})
	}

	t.Run("unsupported type", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "pg"}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "show", "testdb"), `unsupported for service of type "pg"`)
	})

	t.Run("unknown service", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb"}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "show", "nope"), "not found")
	})
}

func TestDBAASAclList(t *testing.T) {
	t.Run("kafka", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "kafka", KafkaAcls: testKafkaAcls()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "list", "testdb"))

		acls := testKafkaAcls()
		require.Equal(t, &dbaasAclListOutput{
			{Username: "alice", Kind: "topic", Resource: "orders", Permission: "read", ID: "topic-1"},
			{Username: "alice", Kind: "topic", Resource: "orders", Permission: "write", ID: "topic-2"},
			{Username: "bob", Kind: "topic", Resource: "*", Permission: "admin", ID: "topic-3"},
			{Username: "alice", Kind: "schema-registry", Resource: "Subject:orders", Permission: "schema_registry_read", ID: "schema-1"},
		}, kafkaAclListOutput(&acls))
	})

	t.Run("opensearch", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "list", "testdb"))

		config := testOpensearchAcl()
		require.Equal(t, &dbaasAclListOutput{
			{Username: "alice", Kind: "index", Resource: "logs-*", Permission: "read"},
			{Username: "alice", Kind: "index", Resource: "metrics-*", Permission: "write"},
			{Username: "bob", Kind: "index", Resource: "*", Permission: "admin"},
		}, opensearchAclListOutput(&config))
	})

	t.Run("unsupported type", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "clickhouse"}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "list", "testdb"), `unsupported for service of type "clickhouse"`)
	})
}

func TestDBAASAclCreate(t *testing.T) {
	t.Run("kafka topic", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "kafka"}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "create", "testdb", "carol", "--kafka-topic", "orders", "--permission", "readwrite"))
		require.Equal(t, &v3.DBAASKafkaTopicAclEntry{
			Username:   "carol",
			Topic:      "orders",
			Permission: v3.DBAASKafkaTopicAclEntryPermissionReadwrite,
		}, opts.KafkaTopicAclCreate)
		require.Nil(t, opts.KafkaSchemaRegistryAclCreate)
	})

	t.Run("kafka schema registry", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "kafka"}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "create", "testdb", "carol",
			"--kafka-schema-registry-resource", "Subject:orders", "--permission", "schema_registry_write"))
		require.Equal(t, &v3.DBAASKafkaSchemaRegistryAclEntry{
			Username:   "carol",
			Resource:   "Subject:orders",
			Permission: v3.DBAASKafkaSchemaRegistryAclEntryPermissionSchemaRegistryWrite,
		}, opts.KafkaSchemaRegistryAclCreate)
		require.Nil(t, opts.KafkaTopicAclCreate)
	})

	t.Run("kafka invalid flags", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "kafka"}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "create", "testdb", "carol", "--permission", "read"), "is required")
		require.ErrorContains(t, runAclCmd(t, "create", "testdb", "carol", "--permission", "read",
			"--kafka-topic", "orders", "--kafka-schema-registry-resource", "Subject:orders"), "mutually exclusive")
		require.ErrorContains(t, runAclCmd(t, "create", "testdb", "carol", "--kafka-topic", "orders"), "--permission is required")
		require.Empty(t, opts.Requests)
	})

	for _, dbType := range []string{"clickhouse", "opensearch"} {
		t.Run("unsupported type "+dbType, func(t *testing.T) {
			opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: dbType}
			ts := setupAclTestServer(t, opts)
			defer ts.Close()

			testutils.SetupV3Client(t, ts.URL)

			require.ErrorContains(t, runAclCmd(t, "create", "testdb", "carol", "--permission", "read"),
				`unsupported for service of type "`+dbType+`"`)
		})
	}
}

func TestDBAASAclDelete(t *testing.T) {
	t.Run("kafka topic", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "kafka", KafkaAcls: testKafkaAcls()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "delete", "testdb", "alice", "--kafka-topic", "orders", "--permission", "write", "--force"))
		require.Equal(t, []string{"GET /acl-config", "DELETE /topic/acl-config/topic-2"}, opts.Requests)
	})

	t.Run("kafka schema registry", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "kafka", KafkaAcls: testKafkaAcls()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "delete", "testdb", "alice", "--kafka-schema-registry-resource", "Subject:orders", "--force"))
		require.Equal(t, []string{"GET /acl-config", "DELETE /schema-registry/acl-config/schema-1"}, opts.Requests)
	})

	t.Run("kafka by id", func(t *testing.T) {
		acls := testKafkaAcls()
		// An exact duplicate, which only its ID can select.
		acls.TopicAcl = append(acls.TopicAcl, v3.DBAASKafkaTopicAclEntry{
			ID: "topic-duplicate", Username: "alice", Topic: "orders", Permission: v3.DBAASKafkaTopicAclEntryPermissionRead,
		})
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "kafka", KafkaAcls: acls}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "delete", "testdb", "alice", "--kafka-topic", "orders", "--permission", "read", "--force"),
			"(topic-1, topic-duplicate), use --permission or --id")

		opts.Requests = nil
		require.NoError(t, runAclCmd(t, "delete", "testdb", "alice", "--id", "topic-duplicate", "--force"))
		require.Equal(t, []string{"GET /acl-config", "DELETE /topic/acl-config/topic-duplicate"}, opts.Requests)

		opts.Requests = nil
		require.NoError(t, runAclCmd(t, "delete", "testdb", "alice", "--id", "schema-1", "--force"))
		require.Equal(t, []string{"GET /acl-config", "DELETE /schema-registry/acl-config/schema-1"}, opts.Requests)

		// The ID must belong to the user given.
		opts.Requests = nil
		require.ErrorContains(t, runAclCmd(t, "delete", "testdb", "alice", "--id", "topic-3", "--force"), "no matching ACL entry")
		require.Equal(t, []string{"GET /acl-config"}, opts.Requests)

		opts.Requests = nil
		require.ErrorContains(t, runAclCmd(t, "delete", "testdb", "alice", "--id", "topic-1", "--kafka-topic", "orders", "--force"),
			"--id can't be used with")
		require.Empty(t, opts.Requests)
	})

	t.Run("kafka no single match", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "kafka", KafkaAcls: testKafkaAcls()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "delete", "testdb", "alice", "--kafka-topic", "orders", "--force"), "use --permission")
		require.ErrorContains(t, runAclCmd(t, "delete", "testdb", "alice", "--kafka-topic", "nope", "--force"), "no matching ACL entry")
		require.ErrorContains(t, runAclCmd(t, "delete", "testdb", "alice", "--force"), "exactly one of")
		for _, r := range opts.Requests {
			require.Equal(t, "GET /acl-config", r)
		}
	})

	t.Run("opensearch unsupported", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "delete", "testdb", "alice", "--force"), `unsupported for service of type "opensearch"`)
		require.Empty(t, opts.Requests)
	})
}
