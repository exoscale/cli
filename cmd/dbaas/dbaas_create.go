package dbaas

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"

	"github.com/exoscale/cli/pkg/output"
)

type dbaasServiceCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Type string `cli-arg:"#"`
	Plan string `cli-arg:"#"`
	Name string `cli-arg:"#"`

	HelpKafka      bool `cli-usage:"show usage for flags specific to the kafka type"`
	HelpOpensearch bool `cli-usage:"show usage for flags specific to the opensearch type"`
	HelpMysql      bool `cli-usage:"show usage for flags specific to the mysql type"`
	HelpPg         bool `cli-usage:"show usage for flags specific to the pg type"`
	HelpValkey     bool `cli-usage:"show usage for flags specific to the valkey type"`
	HelpGrafana    bool `cli-usage:"show usage for flags specific to the grafana type"`

	MaintenanceDOW        string `cli-flag:"maintenance-dow" cli-usage:"automated Database Service maintenance day-of-week"`
	MaintenanceTime       string `cli-usage:"automated Database Service maintenance time (format HH:MM:SS)"`
	TerminationProtection bool   `cli-usage:"enable Database Service termination protection; set --termination-protection=false to disable"`
	Zone                  string `cli-short:"z" cli-usage:"Database Service zone"`

	// "grafana" type specific flags
	GrafanaForkFrom string   `cli-flag:"grafana-fork-from" cli-usage:"name of a Database Service to fork from" cli-hidden:""`
	GrafanaIPFilter []string `cli-flag:"grafana-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	GrafanaSettings string   `cli-flag:"grafana-settings" cli-usage:"Grafana configuration settings (JSON format)" cli-hidden:""`

	// "kafka" type specific flags
	KafkaConnectSettings        string   `cli-flag:"kafka-connect-settings" cli-usage:"Kafka Connect configuration settings (JSON format)" cli-hidden:""`
	KafkaEnableCertAuth         bool     `cli-flag:"kafka-enable-cert-auth" cli-usage:"enable certificate-based authentication method" cli-hidden:""`
	KafkaEnableKafkaConnect     bool     `cli-flag:"kafka-enable-kafka-connect" cli-usage:"enable Kafka Connect" cli-hidden:""`
	KafkaEnableKafkaREST        bool     `cli-flag:"kafka-enable-kafka-rest" cli-usage:"enable Kafka REST" cli-hidden:""`
	KafkaEnableSASLAuth         bool     `cli-flag:"kafka-enable-sasl-auth" cli-usage:"enable SASL-based authentication method" cli-hidden:""`
	KafkaEnableSchemaRegistry   bool     `cli-flag:"kafka-enable-schema-registry" cli-usage:"enable Schema Registry" cli-hidden:""`
	KafkaIPFilter               []string `cli-flag:"kafka-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	KafkaRESTSettings           string   `cli-flag:"kafka-rest-settings" cli-usage:"Kafka REST configuration settings (JSON format)" cli-hidden:""`
	KafkaSchemaRegistrySettings string   `cli-flag:"kafka-schema-registry-settings" cli-usage:"Schema Registry configuration settings (JSON format)" cli-hidden:""`
	KafkaSettings               string   `cli-flag:"kafka-settings" cli-usage:"Kafka configuration settings (JSON format)" cli-hidden:""`
	KafkaVersion                string   `cli-flag:"kafka-version" cli-usage:"Kafka major version" cli-hidden:""`

	// "opensearch" type specific flags
	OpensearchForkFromService                        string   `cli-flag:"opensearch-fork-from-service" cli-usage:"Service name" cli-hidden:""`
	OpensearchIndexPatterns                          string   `cli-flag:"opensearch-index-patterns" cli-usage:"JSON Array of index patterns (https://openapi-v2.exoscale.com/#operation-get-dbaas-service-opensearch-200-index-patterns)" cli-hidden:""`
	OpensearchIndexTemplateMappingNestedObjectsLimit int64    `cli-flag:"opensearch-index-template-mapping-nested-objects-limit" cli-usage:"The maximum number of nested cli-flag objects that a single document can contain across all nested types. Default is 10000." cli-hidden:""`
	OpensearchIndexTemplateNumberOfReplicas          int64    `cli-flag:"opensearch-index-template-number-of-replicas" cli-usage:"The number of replicas each primary shard has." cli-hidden:""`
	OpensearchIndexTemplateNumberOfShards            int64    `cli-flag:"opensearch-index-template-number-of-shards" cli-usage:"The number of primary shards that an index should have." cli-hidden:""`
	OpensearchIPFilter                               []string `cli-flag:"opensearch-ip-filter" cli-usage:"Allow incoming connections from CIDR address block" cli-hidden:""`
	OpensearchKeepIndexRefreshInterval               bool     `cli-flag:"opensearch-keep-index-refresh-interval" cli-usage:"index.refresh_interval is reset to default value for every index to be sure that indices are always visible to search. Set to true disable this." cli-hidden:""`
	OpensearchDashboardEnabled                       bool     `cli-flag:"opensearch-dashboard-enabled" cli-usage:"Enable or disable OpenSearch Dashboards (default: true)" cli-hidden:""`
	OpensearchDashboardMaxOldSpaceSize               int64    `cli-flag:"opensearch-dashboard-max-old-space-size" cli-usage:"Memory limit in MiB for OpenSearch Dashboards. Note: The memory reserved by OpenSearch Dashboards is not available for OpenSearch. (default: 128)" cli-hidden:""`
	OpensearchDashboardRequestTimeout                int64    `cli-flag:"opensearch-dashboard-request-timeout" cli-usage:"Timeout in milliseconds for requests made by OpenSearch Dashboards towards OpenSearch (default: 30000)" cli-hidden:""`
	OpensearchSettings                               string   `cli-flag:"opensearch-settings" cli-usage:"OpenSearch-specific settings (JSON)" cli-hidden:""`
	OpensearchRecoveryBackupName                     string   `cli-flag:"opensearch-recovery-backup-name" cli-usage:"Name of a backup to recover from for services that support backup names" cli-hidden:""`
	OpensearchVersion                                string   `cli-flag:"opensearch-version" cli-usage:"OpenSearch major version" cli-hidden:""`

	// "mysql" type specific flags
	MysqlAdminPassword         string   `cli-flag:"mysql-admin-password" cli-usage:"custom password for admin user" cli-hidden:""`
	MysqlAdminUsername         string   `cli-flag:"mysql-admin-username" cli-usage:"custom username for admin user" cli-hidden:""`
	MysqlBackupSchedule        string   `cli-flag:"mysql-backup-schedule" cli-usage:"automated backup schedule (format: HH:MM)" cli-hidden:""`
	MysqlForkFrom              string   `cli-flag:"mysql-fork-from" cli-usage:"name of a Database Service to fork from" cli-hidden:""`
	MysqlIPFilter              []string `cli-flag:"mysql-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	MysqlRecoveryBackupTime    string   `cli-flag:"mysql-recovery-backup-time" cli-usage:"the timestamp of the backup to restore when forking from a Database Service" cli-hidden:""`
	MysqlSettings              string   `cli-flag:"mysql-settings" cli-usage:"MySQL configuration settings (JSON format)" cli-hidden:""`
	MysqlVersion               string   `cli-flag:"mysql-version" cli-usage:"MySQL major version" cli-hidden:""`
	MysqlMigrationHost         string   `cli-flag:"mysql-migration-host" cli-usage:"hostname or IP address of the source server where to migrate data from" cli-hidden:""`
	MysqlMigrationPort         int64    `cli-flag:"mysql-migration-port" cli-usage:"port number of the source server where to migrate data from" cli-hidden:""`
	MysqlMigrationPassword     string   `cli-flag:"mysql-migration-password" cli-usage:"password for authenticating to the source server" cli-hidden:""`
	MysqlMigrationSSL          bool     `cli-flag:"mysql-migration-ssl" cli-usage:"connect to the source server using SSL" cli-hidden:""`
	MysqlMigrationUsername     string   `cli-flag:"mysql-migration-username" cli-usage:"username for authenticating to the source server" cli-hidden:""`
	MysqlMigrationDBName       string   `cli-flag:"mysql-migration-dbname" cli-usage:"database name for bootstrapping the initial connection" cli-hidden:""`
	MysqlMigrationMethod       string   `cli-flag:"mysql-migration-method" cli-usage:"migration method to be used (\"dump\" or \"replication\")" cli-hidden:""`
	MysqlMigrationIgnoreDbs    []string `cli-flag:"mysql-migration-ignore-dbs" cli-usage:"list of databases which should be ignored during migration" cli-hidden:""`
	MysqlBinlogRetentionPeriod int64    `cli-flag:"mysql-binlog-retention-period" cli-usage:"the minimum amount of time in seconds to keep binlog entries before deletion" cli-hidden:""`

	// "pg" type specific flags
	PGAdminPassword      string   `cli-flag:"pg-admin-password" cli-usage:"custom password for admin user" cli-hidden:""`
	PGAdminUsername      string   `cli-flag:"pg-admin-username" cli-usage:"custom username for admin user" cli-hidden:""`
	PGBackupSchedule     string   `cli-flag:"pg-backup-schedule" cli-usage:"automated backup schedule (format: HH:MM)" cli-hidden:""`
	PGBouncerSettings    string   `cli-flag:"pg-bouncer-settings" cli-usage:"PgBouncer configuration settings (JSON format)" cli-hidden:""`
	PGForkFrom           string   `cli-flag:"pg-fork-from" cli-usage:"name of a Database Service to fork from" cli-hidden:""`
	PGIPFilter           []string `cli-flag:"pg-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	PGLookoutSettings    string   `cli-flag:"pg-lookout-settings" cli-usage:"pglookout configuration settings (JSON format)" cli-hidden:""`
	PGRecoveryBackupTime string   `cli-flag:"pg-recovery-backup-time" cli-usage:"the timestamp of the backup to restore when forking from a Database Service" cli-hidden:""`
	PGSettings           string   `cli-flag:"pg-settings" cli-usage:"PostgreSQL configuration settings (JSON format)" cli-hidden:""`
	PGVersion            string   `cli-flag:"pg-version" cli-usage:"PostgreSQL major version" cli-hidden:""`
	PGMigrationHost      string   `cli-flag:"pg-migration-host" cli-usage:"hostname or IP address of the source server where to migrate data from" cli-hidden:""`
	PGMigrationPort      int64    `cli-flag:"pg-migration-port" cli-usage:"port number of the source server where to migrate data from" cli-hidden:""`
	PGMigrationPassword  string   `cli-flag:"pg-migration-password" cli-usage:"password for authenticating to the source server" cli-hidden:""`
	PGMigrationSSL       bool     `cli-flag:"pg-migration-ssl" cli-usage:"connect to the source server using SSL" cli-hidden:""`
	PGMigrationUsername  string   `cli-flag:"pg-migration-username" cli-usage:"username for authenticating to the source server" cli-hidden:""`
	PGMigrationDBName    string   `cli-flag:"pg-migration-dbname" cli-usage:"database name for bootstrapping the initial connection" cli-hidden:""`
	PGMigrationMethod    string   `cli-flag:"pg-migration-method" cli-usage:"migration method to be used (\"dump\" or \"replication\")" cli-hidden:""`
	PGMigrationIgnoreDbs []string `cli-flag:"pg-migration-ignore-dbs" cli-usage:"list of databases which should be ignored during migration" cli-hidden:""`

	// "valkey" type specific flags
	ValkeyForkFrom           string   `cli-flag:"valkey-fork-from" cli-usage:"name of a Database Service to fork from" cli-hidden:""`
	ValkeyIPFilter           []string `cli-flag:"valkey-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	ValkeyRecoveryBackupName string   `cli-flag:"valkey-recovery-backup-name" cli-usage:"the name of the backup to restore when forking from a Database Service" cli-hidden:""`
	ValkeySettings           string   `cli-flag:"valkey-settings" cli-usage:"Valkey configuration settings (JSON format)" cli-hidden:""`
	ValkeyMigrationHost      string   `cli-flag:"valkey-migration-host" cli-usage:"hostname or IP address of the source server where to migrate data from" cli-hidden:""`
	ValkeyMigrationPort      int64    `cli-flag:"valkey-migration-port" cli-usage:"port number of the source server where to migrate data from" cli-hidden:""`
	ValkeyMigrationPassword  string   `cli-flag:"valkey-migration-password" cli-usage:"password for authenticating to the source server" cli-hidden:""`
	ValkeyMigrationSSL       bool     `cli-flag:"valkey-migration-ssl" cli-usage:"connect to the source server using SSL" cli-hidden:""`
	ValkeyMigrationUsername  string   `cli-flag:"valkey-migration-username" cli-usage:"username for authenticating to the source server" cli-hidden:""`
	ValkeyMigrationDBName    string   `cli-flag:"valkey-migration-dbname" cli-usage:"database name for bootstrapping the initial connection" cli-hidden:""`
	ValkeyMigrationMethod    string   `cli-flag:"valkey-migration-method" cli-usage:"migration method to be used (\"dump\" or \"replication\")" cli-hidden:""`
	ValkeyMigrationIgnoreDbs []string `cli-flag:"valkey-migration-ignore-dbs" cli-usage:"list of databases which should be ignored during migration" cli-hidden:""`
}

func (c *dbaasServiceCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *dbaasServiceCreateCmd) CmdShort() string { return "Create a Database Service" }

func (c *dbaasServiceCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Database Service.

Supported values for --maintenance-dow: %s

Supported output template annotations: %s`,
		strings.Join(dbServiceMaintenanceDOWs, ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceShowOutput{}), ", "))
}

func (c *dbaasServiceCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	switch {
	case cmd.Flags().Changed("help-grafana"):
		exocmd.CmdShowHelpFlags(cmd.Flags(), "grafana-")
		os.Exit(0)
	case cmd.Flags().Changed("help-kafka"):
		exocmd.CmdShowHelpFlags(cmd.Flags(), "kafka-")
		os.Exit(0)
	case cmd.Flags().Changed("help-opensearch"):
		exocmd.CmdShowHelpFlags(cmd.Flags(), "opensearch-")
		os.Exit(0)
	case cmd.Flags().Changed("help-mysql"):
		exocmd.CmdShowHelpFlags(cmd.Flags(), "mysql-")
		os.Exit(0)
	case cmd.Flags().Changed("help-pg"):
		exocmd.CmdShowHelpFlags(cmd.Flags(), "pg-")
		os.Exit(0)
	case cmd.Flags().Changed("help-valkey"):
		exocmd.CmdShowHelpFlags(cmd.Flags(), "valkey-")
		os.Exit(0)
	}

	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceCreateCmd) CmdRun(cmd *cobra.Command, args []string) error {
	if (cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceDOW)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceTime))) &&
		(!cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceDOW)) ||
			!cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceTime))) {
		return fmt.Errorf(
			"both --%s and --%s must be specified",
			exocmd.MustCLICommandFlagName(c, &c.MaintenanceDOW),
			exocmd.MustCLICommandFlagName(c, &c.MaintenanceTime))
	}

	switch c.Type {
	case "grafana":
		return c.createGrafana(cmd, args)
	case "kafka":
		return c.createKafka(cmd, args)
	case "opensearch":
		return c.createOpensearch(cmd, args)
	case "mysql":
		return c.createMysql(cmd, args)
	case "pg":
		return c.createPG(cmd, args)
	case "valkey":
		return c.createValkey(cmd, args)
	default:
		return fmt.Errorf("unsupported service type %q", c.Type)
	}
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasCmd, &dbaasServiceCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		TerminationProtection: true,
	}))
}
