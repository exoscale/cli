package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type dbServiceCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Type string `cli-arg:"#"`
	Plan string `cli-arg:"#"`
	Name string `cli-arg:"#"`

	ForkFrom              string `cli-usage:"name of a Database Service to fork from"`
	MaintenanceDOW        string `cli-flag:"maintenance-dow" cli-usage:"automated Database Service maintenance day-of-week"`
	MaintenanceTime       string `cli-usage:"automated Database Service maintenance time (format HH:MM:SS)"`
	TerminationProtection bool   `cli-usage:"enable Database Service termination protection"`
	Zone                  string `cli-short:"z" cli-usage:"Database Service zone"`

	// "kafka" type specific flags
	KafkaConnectSettings        string   `cli-flag:"kafka-connect-settings" cli-usage:"[kafka] Kafka Connect configuration settings (JSON format)"`
	KafkaEnableCertAuth         bool     `cli-flag:"kafka-enable-cert-auth" cli-usage:"[kafka] enable certificate-based authentication method"`
	KafkaEnableKafkaConnect     bool     `cli-flag:"kafka-enable-kafka-connect" cli-usage:"[kafka] enable Kafka Connect"`
	KafkaEnableKafkaREST        bool     `cli-flag:"kafka-enable-kafka-rest" cli-usage:"[kafka] enable Kafka REST"`
	KafkaEnableSASLAuth         bool     `cli-flag:"kafka-enable-sasl-auth" cli-usage:"[kafka] enable SASL-based authentication method"`
	KafkaEnableSchemaRegistry   bool     `cli-flag:"kafka-enable-schema-registry" cli-usage:"[kafka] enable Schema Registry"`
	KafkaIPFilter               []string `cli-flag:"kafka-ip-filter" cli-usage:"[kafka] allow incoming connections from CIDR address block"`
	KafkaRESTSettings           string   `cli-flag:"kafka-rest-settings" cli-usage:"[kafka] Kafka REST configuration settings (JSON format)"`
	KafkaSchemaRegistrySettings string   `cli-flag:"kafka-schema-registry-settings" cli-usage:"[kafka] Schema Registry configuration settings (JSON format)"`
	KafkaSettings               string   `cli-flag:"kafka-settings" cli-usage:"[kafka] Kafka configuration settings (JSON format)"`
	KafkaVersion                string   `cli-flag:"kafka-version" cli-usage:"[kafka] Kafka major version"`

	// "mysql" type specific flags
	MysqlAdminPassword  string   `cli-flag:"mysql-admin-password" cli-usage:"[mysql] custom password for admin user"`
	MysqlAdminUsername  string   `cli-flag:"mysql-admin-username" cli-usage:"[mysql] custom username for admin user"`
	MysqlBackupSchedule string   `cli-flag:"mysql-backup-schedule" cli-usage:"[mysql] automated backup schedule (format: HH:MM)"`
	MysqlIPFilter       []string `cli-flag:"mysql-ip-filter" cli-usage:"[mysql] allow incoming connections from CIDR address block"`
	MysqlSettings       string   `cli-flag:"mysql-settings" cli-usage:"[mysql] MySQL configuration settings (JSON format)"`
	MysqlVersion        string   `cli-flag:"mysql-version" cli-usage:"[mysql] MySQL major version"`

	// "pg" type specific flags
	PGAdminPassword   string   `cli-flag:"pg-admin-password" cli-usage:"[pg] custom password for admin user"`
	PGAdminUsername   string   `cli-flag:"pg-admin-username" cli-usage:"[pg] custom username for admin user"`
	PGBackupSchedule  string   `cli-flag:"pg-backup-schedule" cli-usage:"[pg] automated backup schedule (format: HH:MM)"`
	PGBouncerSettings string   `cli-flag:"pg-bouncer-settings" cli-usage:"[pg] PgBouncer configuration settings (JSON format)"`
	PGIPFilter        []string `cli-flag:"pg-ip-filter" cli-usage:"[pg] allow incoming connections from CIDR address block"`
	PGLookoutSettings string   `cli-flag:"pg-lookout-settings" cli-usage:"[pg] pglookout configuration settings (JSON format)"`
	PGSettings        string   `cli-flag:"pg-settings" cli-usage:"[pg] PostgreSQL configuration settings (JSON format)"`
	PGVersion         string   `cli-flag:"pg-version" cli-usage:"[pg] PostgreSQL major version"`

	// "redis" type specific flags
	RedisIPFilter []string `cli-flag:"redis-ip-filter" cli-usage:"[redis] allow incoming connections from CIDR address block"`
	RedisSettings string   `cli-flag:"redis-settings" cli-usage:"[redis] PostgreSQL configuration settings (JSON format)"`
}

func (c *dbServiceCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *dbServiceCreateCmd) cmdShort() string { return "Create a Database Service" }

func (c *dbServiceCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Database Service.

Supported values for --maintenance-dow: %s

Supported output template annotations: %s`,
		strings.Join(dbServiceMaintenanceDOWs, ", "),
		strings.Join(outputterTemplateAnnotations(&dbServiceShowOutput{}), ", "))
}

func (c *dbServiceCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbServiceCreateCmd) cmdRun(cmd *cobra.Command, args []string) error {
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
	cobra.CheckErr(registerCLICommand(dbCmd, &dbServiceCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		TerminationProtection: true,
	}))
}
