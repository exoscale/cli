package cmd

import (
	"fmt"
	"os"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasServiceUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Name string `cli-arg:"#"`

	HelpKafka             bool   `cli-usage:"show usage for flags specific to the kafka type"`
	HelpMysql             bool   `cli-usage:"show usage for flags specific to the mysql type"`
	HelpPg                bool   `cli-usage:"show usage for flags specific to the pg type"`
	HelpRedis             bool   `cli-usage:"show usage for flags specific to the redis type"`
	MaintenanceDOW        string `cli-flag:"maintenance-dow" cli-usage:"automated Database Service maintenance day-of-week"`
	MaintenanceTime       string `cli-usage:"automated Database Service maintenance time (format HH:MM:SS)"`
	Plan                  string `cli-usage:"Database Service plan"`
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

	// "mysql" type specific flags
	MysqlBackupSchedule        string   `cli-flag:"mysql-backup-schedule" cli-usage:"automated backup schedule (format: HH:MM)" cli-hidden:""`
	MysqlIPFilter              []string `cli-flag:"mysql-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	MysqlSettings              string   `cli-flag:"mysql-settings" cli-usage:"MySQL configuration settings (JSON format)" cli-hidden:""`
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
	PGBackupSchedule     string   `cli-flag:"pg-backup-schedule" cli-usage:"automated backup schedule (format: HH:MM)" cli-hidden:""`
	PGBouncerSettings    string   `cli-flag:"pg-bouncer-settings" cli-usage:"PgBouncer configuration settings (JSON format)" cli-hidden:""`
	PGIPFilter           []string `cli-flag:"pg-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	PGLookoutSettings    string   `cli-flag:"pg-lookout-settings" cli-usage:"pglookout configuration settings (JSON format)" cli-hidden:""`
	PGSettings           string   `cli-flag:"pg-settings" cli-usage:"PostgreSQL configuration settings (JSON format)" cli-hidden:""`
	PGMigrationHost      string   `cli-flag:"pg-migration-host" cli-usage:"hostname or IP address of the source server where to migrate data from" cli-hidden:""`
	PGMigrationPort      int64    `cli-flag:"pg-migration-port" cli-usage:"port number of the source server where to migrate data from" cli-hidden:""`
	PGMigrationPassword  string   `cli-flag:"pg-migration-password" cli-usage:"password for authenticating to the source server" cli-hidden:""`
	PGMigrationSSL       bool     `cli-flag:"pg-migration-ssl" cli-usage:"connect to the source server using SSL" cli-hidden:""`
	PGMigrationUsername  string   `cli-flag:"pg-migration-username" cli-usage:"username for authenticating to the source server" cli-hidden:""`
	PGMigrationDbName    string   `cli-flag:"pg-migration-dbname" cli-usage:"database name for bootstrapping the initial connection" cli-hidden:""`
	PGMigrationMethod    string   `cli-flag:"pg-migration-method" cli-usage:"migration method to be used (\"dump\" or \"replication\")" cli-hidden:""`
	PGMigrationIgnoreDbs []string `cli-flag:"pg-migration-ignore-dbs" cli-usage:"list of databases which should be ignored during migration" cli-hidden:""`

	// "redis" type specific flags
	RedisIPFilter           []string `cli-flag:"redis-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
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

func (c *dbaasServiceUpdateCmd) cmdAliases() []string { return nil }

func (c *dbaasServiceUpdateCmd) cmdShort() string { return "Update Database Service" }

func (c *dbaasServiceUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates a Database Service.

Supported values for --maintenance-dow: %s

Supported output template annotations: %s`,
		strings.Join(dbServiceMaintenanceDOWs, ", "),
		strings.Join(outputterTemplateAnnotations(&dbServiceShowOutput{}), ", "),
	)
}

func (c *dbaasServiceUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
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

func (c *dbaasServiceUpdateCmd) cmdRun(cmd *cobra.Command, args []string) error {
	if (cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) ||
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime))) &&
		(!cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) ||
			!cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime))) {
		return fmt.Errorf(
			"both --%s and --%s must be specified",
			mustCLICommandFlagName(c, &c.MaintenanceDOW),
			mustCLICommandFlagName(c, &c.MaintenanceTime))
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	databaseServices, err := cs.ListDatabaseServices(ctx, c.Zone)
	if err != nil {
		return err
	}

	var (
		ok              bool
		databaseService *egoscale.DatabaseService
	)
	for _, databaseService = range databaseServices {
		if *databaseService.Name == c.Name {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("%q Database Service not found", c.Name)
	}

	switch *databaseService.Type {
	case "kafka":
		return c.updateKafka(cmd, args)
	case "mysql":
		return c.updateMysql(cmd, args)
	case "pg":
		return c.updatePG(cmd, args)
	case "redis":
		return c.updateRedis(cmd, args)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasCmd, &dbaasServiceUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
