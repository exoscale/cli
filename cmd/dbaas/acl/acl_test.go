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

	KafkaAcls     v3.DBAASKafkaAcls
	OpensearchAcl v3.DBAASOpensearchAclConfig

	// Requests received by the server, as "METHOD PATH".
	Requests []string
	// Last OpenSearch ACL configuration sent to the server.
	OpensearchAclUpdate *v3.DBAASOpensearchAclConfig
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
		switch r.Method {
		case http.MethodGet:
			testutils.WriteJSON(t, w, http.StatusOK, opts.OpensearchAcl)
		case http.MethodPut:
			opts.OpensearchAclUpdate = &v3.DBAASOpensearchAclConfig{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(opts.OpensearchAclUpdate))
			writeOperation(w)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
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
		&dbaasAclUpdateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()},
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

	t.Run("opensearch new user", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "create", "testdb", "carol", "--opensearch-index", "logs-*", "--permission", "deny"))

		want := testOpensearchAcl()
		want.Acls = append(want.Acls, v3.DBAASOpensearchAclConfigAcls{
			Username: "carol",
			Rules:    []v3.DBAASOpensearchAclConfigAclsRules{{Index: "logs-*", Permission: v3.EnumOpensearchRulePermissionDeny}},
		})
		require.Equal(t, &want, opts.OpensearchAclUpdate)
	})

	t.Run("opensearch existing user", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "create", "testdb", "bob", "--opensearch-index", "logs-*", "--permission", "read"))

		want := testOpensearchAcl()
		want.Acls[1].Rules = append(want.Acls[1].Rules,
			v3.DBAASOpensearchAclConfigAclsRules{Index: "logs-*", Permission: v3.EnumOpensearchRulePermissionRead})
		require.Equal(t, &want, opts.OpensearchAclUpdate)
	})

	t.Run("opensearch existing entry", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t,
			runAclCmd(t, "create", "testdb", "alice", "--opensearch-index", "logs-*", "--permission", "write"),
			"already exists")
		require.ErrorContains(t, runAclCmd(t, "create", "testdb", "alice", "--permission", "write"), "--opensearch-index is required")
		require.Nil(t, opts.OpensearchAclUpdate)
	})

	t.Run("unsupported type", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "clickhouse"}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "create", "testdb", "carol", "--permission", "read"),
			`unsupported for service of type "clickhouse"`)
	})
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

	t.Run("opensearch user", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "delete", "testdb", "alice", "--force"))

		want := testOpensearchAcl()
		want.Acls = want.Acls[1:]
		require.Equal(t, &want, opts.OpensearchAclUpdate)
	})

	t.Run("opensearch index", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "delete", "testdb", "alice", "--opensearch-index", "logs-*", "--force"))

		want := testOpensearchAcl()
		want.Acls[0].Rules = want.Acls[0].Rules[1:]
		require.Equal(t, &want, opts.OpensearchAclUpdate)
	})

	t.Run("opensearch last index of a user", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "delete", "testdb", "bob", "--opensearch-index", "*", "--force"))

		want := testOpensearchAcl()
		want.Acls = want.Acls[:1]
		require.Equal(t, &want, opts.OpensearchAclUpdate)
	})

	t.Run("opensearch no match", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "delete", "testdb", "carol", "--force"), "no matching ACL entry")
		require.ErrorContains(t, runAclCmd(t, "delete", "testdb", "alice", "--opensearch-index", "nope", "--force"), "no matching ACL entry")
		require.Nil(t, opts.OpensearchAclUpdate)
	})
}

func TestDBAASAclUpdate(t *testing.T) {
	t.Run("opensearch permission", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "update", "testdb", "alice", "--opensearch-index", "logs-*", "--permission", "readwrite"))

		want := testOpensearchAcl()
		want.Acls[0].Rules[0].Permission = v3.EnumOpensearchRulePermissionReadwrite
		require.Equal(t, &want, opts.OpensearchAclUpdate)
	})

	t.Run("opensearch toggles", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.NoError(t, runAclCmd(t, "update", "testdb", "--opensearch-acl-enabled=false", "--opensearch-extended-acl-enabled"))

		want := testOpensearchAcl()
		disabled, enabled := false, true
		want.AclEnabled = &disabled
		want.ExtendedAclEnabled = &enabled
		require.Equal(t, &want, opts.OpensearchAclUpdate)
	})

	t.Run("opensearch invalid", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "opensearch", OpensearchAcl: testOpensearchAcl()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "update", "testdb"), "nothing to update")
		require.ErrorContains(t, runAclCmd(t, "update", "testdb", "alice", "--permission", "read"), "must be specified together")
		require.ErrorContains(t,
			runAclCmd(t, "update", "testdb", "alice", "--opensearch-index", "nope", "--permission", "read"),
			"no ACL entry found")
		require.Nil(t, opts.OpensearchAclUpdate)
	})

	t.Run("kafka", func(t *testing.T) {
		opts := &aclTestServerOpts{ServiceName: "testdb", ServiceType: "kafka", KafkaAcls: testKafkaAcls()}
		ts := setupAclTestServer(t, opts)
		defer ts.Close()

		testutils.SetupV3Client(t, ts.URL)

		require.ErrorContains(t, runAclCmd(t, "update", "testdb", "--opensearch-acl-enabled"), `unsupported for service of type "kafka"`)
		require.Empty(t, opts.Requests)
	})
}
