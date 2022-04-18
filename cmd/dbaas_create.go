package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type dbaasServiceCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Type string `cli-arg:"#"`
	Plan string `cli-arg:"#"`
	Name string `cli-arg:"#"`

	HelpKafka             bool   `cli-usage:"show usage for flags specific to the kafka type"`
	HelpMysql             bool   `cli-usage:"show usage for flags specific to the mysql type"`
	HelpPg                bool   `cli-usage:"show usage for flags specific to the pg type"`
	HelpRedis             bool   `cli-usage:"show usage for flags specific to the redis type"`
	MaintenanceDOW        string `cli-flag:"maintenance-dow" cli-usage:"automated Database Service maintenance day-of-week"`
	MaintenanceTime       string `cli-usage:"automated Database Service maintenance time (format HH:MM:SS)"`
	TerminationProtection bool   `cli-usage:"enable Database Service termination protection; set --termination-protection=false to disable"`
	Zone                  string `cli-short:"z" cli-usage:"Database Service zone"`

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

	// "mysql" type specific flags
	MysqlAdminPassword         string   `cli-flag:"mysql-admin-password" cli-usage:"custom password for admin user" cli-hidden:""`
	MysqlAdminUsername         string   `cli-flag:"mysql-admin-username" cli-usage:"custom username for admin user" cli-hidden:""`
	MysqlBackupSchedule        string   `cli-flag:"mysql-backup-schedule" cli-usage:"automated backup schedule (format: HH:MM)" cli-hidden:""`
	MysqlForkFrom              string   `cli-flag:"mysql-fork-from" cli-usage:"name of a Database Service to fork from"`
	MysqlIPFilter              []string `cli-flag:"mysql-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	MysqlRecoveryBackupTime    string   `cli-flag:"mysql-recovery-backup-time" cli-usage:"the timestamp of the backup to restore when forking from a Database Service"`
	MysqlSettings              string   `cli-flag:"mysql-settings" cli-usage:"MySQL configuration settings (JSON format)" cli-hidden:""`
	MysqlVersion               string   `cli-flag:"mysql-version" cli-usage:"MySQL major version" cli-hidden:""`
	MysqlMigrationHost         string   `cli-flag:"mysql-migration-host" cli-usage:"hostname or IP address of the source server where to migrate data from" cli-hidden:""`
	MysqlMigrationPort         int64    `cli-flag:"mysql-migration-port" cli-usage:"port number of the source server where to migrate data from" cli-hidden:""`
	MysqlMigrationPassword     string   `cli-flag:"mysql-migration-password" cli-usage:"password for authenticating to the source server" cli-hidden:""`
	MysqlMigrationSSL          bool     `cli-flag:"mysql-migration-ssl" cli-usage:"connect to the source server using SSL" cli-hidden:""`
	MysqlMigrationUsername     string   `cli-flag:"mysql-migration-username" cli-usage:"username for authenticating to the source server" cli-hidden:""`
	MysqlMigrationDbName       string   `cli-flag:"mysql-migration-dbname" cli-usage:"database name for bootstrapping the initial connection" cli-hidden:""`
	MysqlMigrationMethod       string   `cli-flag:"mysql-migration-method" cli-usage:"migration method to be used (\"dump\" or \"replication\")" cli-hidden:""`
	MysqlMigrationIgnoreDbs    []string `cli-flag:"mysql-migration-ignore-dbs" cli-usage:"list of databases which should be ignored during migration" cli-hidden:""`
	MysqlBinlogRetentionPeriod int64    `cli-flag:"mysql-binlog-retention-period" cli-usage:"the minimum amount of time in seconds to keep binlog entries before deletion" cli-hidden:""`

	// "pg" type specific flags
	PGAdminPassword      string   `cli-flag:"pg-admin-password" cli-usage:"custom password for admin user" cli-hidden:""`
	PGAdminUsername      string   `cli-flag:"pg-admin-username" cli-usage:"custom username for admin user" cli-hidden:""`
	PGBackupSchedule     string   `cli-flag:"pg-backup-schedule" cli-usage:"automated backup schedule (format: HH:MM)" cli-hidden:""`
	PGBouncerSettings    string   `cli-flag:"pg-bouncer-settings" cli-usage:"PgBouncer configuration settings (JSON format)" cli-hidden:""`
	PGForkFrom           string   `cli-flag:"pg-fork-from" cli-usage:"name of a Database Service to fork from"`
	PGIPFilter           []string `cli-flag:"pg-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	PGLookoutSettings    string   `cli-flag:"pg-lookout-settings" cli-usage:"pglookout configuration settings (JSON format)" cli-hidden:""`
	PGRecoveryBackupTime string   `cli-flag:"pg-recovery-backup-time" cli-usage:"the timestamp of the backup to restore when forking from a Database Service"`
	PGSettings           string   `cli-flag:"pg-settings" cli-usage:"PostgreSQL configuration settings (JSON format)" cli-hidden:""`
	PGVersion            string   `cli-flag:"pg-version" cli-usage:"PostgreSQL major version" cli-hidden:""`
	PGMigrationHost      string   `cli-flag:"pg-migration-host" cli-usage:"hostname or IP address of the source server where to migrate data from" cli-hidden:""`
	PGMigrationPort      int64    `cli-flag:"pg-migration-port" cli-usage:"port number of the source server where to migrate data from" cli-hidden:""`
	PGMigrationPassword  string   `cli-flag:"pg-migration-password" cli-usage:"password for authenticating to the source server" cli-hidden:""`
	PGMigrationSSL       bool     `cli-flag:"pg-migration-ssl" cli-usage:"connect to the source server using SSL" cli-hidden:""`
	PGMigrationUsername  string   `cli-flag:"pg-migration-username" cli-usage:"username for authenticating to the source server" cli-hidden:""`
	PGMigrationDbName    string   `cli-flag:"pg-migration-dbname" cli-usage:"database name for bootstrapping the initial connection" cli-hidden:""`
	PGMigrationMethod    string   `cli-flag:"pg-migration-method" cli-usage:"migration method to be used (\"dump\" or \"replication\")" cli-hidden:""`
	PGMigrationIgnoreDbs []string `cli-flag:"pg-migration-ignore-dbs" cli-usage:"list of databases which should be ignored during migration" cli-hidden:""`

	// "redis" type specific flags
	RedisForkFrom           string   `cli-flag:"redis-fork-from" cli-usage:"name of a Database Service to fork from"`
	RedisIPFilter           []string `cli-flag:"redis-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	RedisRecoveryBackupName string   `cli-flag:"redis-recovery-backup-name" cli-usage:"the name of the backup to restore when forking from a Database Service"`
	RedisSettings           string   `cli-flag:"redis-settings" cli-usage:"Redis configuration settings (JSON format)" cli-hidden:""`
	RedisMigrationHost      string   `cli-flag:"redis-migration-host" cli-usage:"hostname or IP address of the source server where to migrate data from" cli-hidden:""`
	RedisMigrationPort      int64    `cli-flag:"redis-migration-port" cli-usage:"port number of the source server where to migrate data from" cli-hidden:""`
	RedisMigrationPassword  string   `cli-flag:"redis-migration-password" cli-usage:"password for authenticating to the source server" cli-hidden:""`
	RedisMigrationSSL       bool     `cli-flag:"redis-migration-ssl" cli-usage:"connect to the source server using SSL" cli-hidden:""`
	RedisMigrationUsername  string   `cli-flag:"redis-migration-username" cli-usage:"username for authenticating to the source server" cli-hidden:""`
	RedisMigrationDbName    string   `cli-flag:"redis-migration-dbname" cli-usage:"database name for bootstrapping the initial connection" cli-hidden:""`
	RedisMigrationMethod    string   `cli-flag:"redis-migration-method" cli-usage:"migration method to be used (\"dump\" or \"replication\")" cli-hidden:""`
	RedisMigrationIgnoreDbs []string `cli-flag:"redis-migration-ignore-dbs" cli-usage:"list of databases which should be ignored during migration" cli-hidden:""`
}

func (c *dbaasServiceCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *dbaasServiceCreateCmd) cmdShort() string { return "Create a Database Service" }

func (c *dbaasServiceCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Database Service.

Supported values for --maintenance-dow: %s

Supported output template annotations: %s`,
		strings.Join(dbServiceMaintenanceDOWs, ", "),
		strings.Join(outputterTemplateAnnotations(&dbServiceShowOutput{}), ", "))
}

func (c *dbaasServiceCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	switch {
	case cmd.Flags().Changed("help-kafka"):
		cmdShowHelpFlags(cmd.Flags(), "kafka-")
		os.Exit(0)
	case cmd.Flags().Changed("help-mysql"):
		cmdShowHelpFlags(cmd.Flags(), "mysql-")
		os.Exit(0)
	case cmd.Flags().Changed("help-pg"):
		cmdShowHelpFlags(cmd.Flags(), "pg-")
		os.Exit(0)
	case cmd.Flags().Changed("help-redis"):
		cmdShowHelpFlags(cmd.Flags(), "redis-")
		os.Exit(0)
	}

	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceCreateCmd) cmdRun(cmd *cobra.Command, args []string) error {
	if (cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) ||
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime))) &&
		(!cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) ||
			!cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime))) {
		return fmt.Errorf(
			"both --%s and --%s must be specified",
			mustCLICommandFlagName(c, &c.MaintenanceDOW),
			mustCLICommandFlagName(c, &c.MaintenanceTime))
	}

	switch c.Type {
	case "kafka":
		return c.createKafka(cmd, args)
	case "mysql":
		return c.createMysql(cmd, args)
	case "pg":
		return c.createPG(cmd, args)
	case "redis":
		return c.createRedis(cmd, args)
	default:
		return fmt.Errorf("unsupported service type %q", c.Type)
	}
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasCmd, &dbaasServiceCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		TerminationProtection: true,
	}))
}
