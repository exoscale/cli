package cmd

import (
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbServiceUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Name string `cli-arg:"#"`

	MaintenanceDOW        string `cli-flag:"maintenance-dow" cli-usage:"automated Database Service maintenance day-of-week"`
	MaintenanceTime       string `cli-usage:"automated Database Service maintenance time (format HH:MM:SS)"`
	Plan                  string `cli-usage:"Database Service plan"`
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

	// "mysql" type specific flags
	MysqlBackupSchedule string   `cli-flag:"mysql-backup-schedule" cli-usage:"[mysql] automated backup schedule (format: HH:MM)"`
	MysqlIPFilter       []string `cli-flag:"mysql-ip-filter" cli-usage:"[mysql] allow incoming connections from CIDR address block"`
	MysqlSettings       string   `cli-flag:"mysql-settings" cli-usage:"[mysql] MySQL configuration settings (JSON format)"`

	// "pg" type specific flags
	PGBackupSchedule  string   `cli-flag:"pg-backup-schedule" cli-usage:"[pg] automated backup schedule (format: HH:MM)"`
	PGBouncerSettings string   `cli-flag:"pg-bouncer-settings" cli-usage:"[pg] PgBouncer configuration settings (JSON format)"`
	PGIPFilter        []string `cli-flag:"pg-ip-filter" cli-usage:"[pg] allow incoming connections from CIDR address block"`
	PGLookoutSettings string   `cli-flag:"pg-lookout-settings" cli-usage:"[pg] pglookout configuration settings (JSON format)"`
	PGSettings        string   `cli-flag:"pg-settings" cli-usage:"[pg] PostgreSQL configuration settings (JSON format)"`

	// "redis" type specific flags
	RedisIPFilter []string `cli-flag:"redis-ip-filter" cli-usage:"[redis] allow incoming connections from CIDR address block"`
	RedisSettings string   `cli-flag:"redis-settings" cli-usage:"[redis] Redis configuration settings (JSON format)"`
}

func (c *dbServiceUpdateCmd) cmdAliases() []string { return nil }

func (c *dbServiceUpdateCmd) cmdShort() string { return "Update Database Service" }

func (c *dbServiceUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates a Database Service.

Supported values for --maintenance-dow: %s

Supported output template annotations: %s`,
		strings.Join(dbServiceMaintenanceDOWs, ", "),
		strings.Join(outputterTemplateAnnotations(&dbServiceShowOutput{}), ", "),
	)
}

func (c *dbServiceUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbServiceUpdateCmd) cmdRun(cmd *cobra.Command, args []string) error {
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
	cobra.CheckErr(registerCLICommand(dbCmd, &dbServiceUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
