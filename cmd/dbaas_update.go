package cmd

import (
	"fmt"
	"os"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbServiceUpdateCmd struct {
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
	TerminationProtection bool   `cli-usage:"enable Database Service termination protection"`
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
	MysqlBackupSchedule string   `cli-flag:"mysql-backup-schedule" cli-usage:"automated backup schedule (format: HH:MM)" cli-hidden:""`
	MysqlIPFilter       []string `cli-flag:"mysql-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	MysqlSettings       string   `cli-flag:"mysql-settings" cli-usage:"MySQL configuration settings (JSON format)" cli-hidden:""`

	// "pg" type specific flags
	PGBackupSchedule  string   `cli-flag:"pg-backup-schedule" cli-usage:"automated backup schedule (format: HH:MM)" cli-hidden:""`
	PGBouncerSettings string   `cli-flag:"pg-bouncer-settings" cli-usage:"PgBouncer configuration settings (JSON format)" cli-hidden:""`
	PGIPFilter        []string `cli-flag:"pg-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	PGLookoutSettings string   `cli-flag:"pg-lookout-settings" cli-usage:"pglookout configuration settings (JSON format)" cli-hidden:""`
	PGSettings        string   `cli-flag:"pg-settings" cli-usage:"PostgreSQL configuration settings (JSON format)" cli-hidden:""`

	// "redis" type specific flags
	RedisIPFilter []string `cli-flag:"redis-ip-filter" cli-usage:"allow incoming connections from CIDR address block" cli-hidden:""`
	RedisSettings string   `cli-flag:"redis-settings" cli-usage:"Redis configuration settings (JSON format)" cli-hidden:""`
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
	cobra.CheckErr(registerCLICommand(dbaasCmd, &dbServiceUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
