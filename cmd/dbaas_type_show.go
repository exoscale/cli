package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exo "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasTypePlanListItemOutput struct {
	Name       string `json:"name"`
	Nodes      int64  `json:"nodes"`
	NodeCPUs   int64  `json:"node_cpus"`
	NodeMemory int64  `json:"node_memory"`
	DiskSpace  int64  `json:"disk_space"`
	Authorized bool   `json:"authorized"`
}

type dbaasTypePlanListOutput []dbaasTypePlanListItemOutput

func (o *dbaasTypePlanListOutput) ToJSON() { output.JSON(o) }
func (o *dbaasTypePlanListOutput) ToText() { output.Text(o) }
func (o *dbaasTypePlanListOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Name", "# Nodes", "# CPUs", "Node Memory", "Disk Space", "Authorized"})
	defer t.Render()

	for _, p := range *o {
		t.Append([]string{
			p.Name,
			fmt.Sprint(p.Nodes),
			fmt.Sprint(p.NodeCPUs),
			humanize.Bytes(uint64(p.NodeMemory)),
			humanize.Bytes(uint64(p.DiskSpace)),
			fmt.Sprint(p.Authorized),
		})
	}
}

type dbaasTypePlanBackupOutput struct {
	Interval                   *int64  `json:"interval"`
	MaxCount                   *int64  `json:"max_count"`
	RecoveryMode               *string `json:"recovery_mode"`
	FrequentIntervalMinutes    *int64  `json:"frequent_interval_minutes"`
	FrequentOldestAgeMinutes   *int64  `json:"frequent_oldest_age_minutes"`
	InfrequentIntervalMinutes  *int64  `json:"infrequent_interval_minutes"`
	InfrequentOldestAgeMinutes *int64  `json:"infrequent_oldest_age_minutes"`
}

func (o *dbaasTypePlanBackupOutput) ToJSON() { output.JSON(o) }
func (o *dbaasTypePlanBackupOutput) ToText() { output.Text(o) }
func (o *dbaasTypePlanBackupOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.Append([]string{"Backup interval (hours)", Int64PtrFormatOutput(o.Interval)})
	t.Append([]string{"Max backups", Int64PtrFormatOutput(o.MaxCount)})
	t.Append([]string{"Recovery mode", utils.DefaultString(o.RecoveryMode, "")})
	t.Append([]string{"Frequent backup interval", Int64PtrFormatOutput(o.FrequentIntervalMinutes)})
	t.Append([]string{"Frequent backup max age", Int64PtrFormatOutput(o.FrequentOldestAgeMinutes)})
	t.Append([]string{"Infrequent backup interval", Int64PtrFormatOutput(o.InfrequentIntervalMinutes)})
	t.Append([]string{"Infrequent backup max age", Int64PtrFormatOutput(o.InfrequentOldestAgeMinutes)})
}

type dbaasTypeShowOutput struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	AvailableVersions []string `json:"available_versions"`
	DefaultVersion    string   `json:"default_version"`
}

func (o *dbaasTypeShowOutput) ToJSON() { output.JSON(o) }
func (o *dbaasTypeShowOutput) ToText() { output.Text(o) }
func (o *dbaasTypeShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Available Versions", strings.Join(o.AvailableVersions, ", ")})
	t.Append([]string{"Default Version", o.DefaultVersion})
}

var (
	grafanaSettings    = []string{"grafana"}
	opensearchSettings = []string{"opensearch"}
	kafkaSettings      = []string{
		"kafka",
		"kafka-rest",
		"kafka-connect",
		"schema-registry",
	}
	mysqlSettings = []string{"mysql"}
	pgSettings    = []string{
		"pg",
		"pgbouncer",
		"pglookout",
	}
	redisSettings = []string{"redis"}
)

type dbaasTypeShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string `cli-arg:"#"`

	ShowPlans        bool   `cli-flag:"plans" cli-usage:"list plans offered for the Database Service type"`
	ShowSettings     string `cli-flag:"settings" cli-usage:"show settings supported by the Database Service type"`
	ShowBackupConfig string `cli-flag:"backup-config" cli-usage:"show backup configuration for the Database Service type and Plan"`
}

func (c *dbaasTypeShowCmd) cmdAliases() []string { return gShowAlias }

func (c *dbaasTypeShowCmd) cmdShort() string { return "Show a Database Service type details" }

func (c *dbaasTypeShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Database Service type details.

Supported Database Service type settings:

* %s
* %s
* %s
* %s
* %s
* %s

Supported output template annotations:

* When showing a Database Service: %s

* When listing Database Service plans: %s`,
		strings.Join(grafanaSettings, ", "),
		strings.Join(opensearchSettings, ", "),
		strings.Join(kafkaSettings, ", "),
		strings.Join(mysqlSettings, ", "),
		strings.Join(pgSettings, ", "),
		strings.Join(redisSettings, ", "),
		strings.Join(output.TemplateAnnotations(&dbaasTypeShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasTypePlanListItemOutput{}), ", "))
}

func (c *dbaasTypeShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasTypeShowCmd) cmdRun(_ *cobra.Command, _ []string) error { //nolint:gocyclo
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone),
	)

	dt, err := globalstate.EgoscaleClient.GetDatabaseServiceType(ctx, account.CurrentAccount.DefaultZone, c.Name)
	if err != nil {
		return err
	}

	if c.ShowPlans {
		out := make(dbaasTypePlanListOutput, len(dt.Plans))
		for i := range dt.Plans {
			out[i] = dbaasTypePlanListItemOutput{
				Name:       *dt.Plans[i].Name,
				Nodes:      *dt.Plans[i].Nodes,
				NodeCPUs:   *dt.Plans[i].NodeCPUs,
				NodeMemory: *dt.Plans[i].NodeMemory,
				DiskSpace:  *dt.Plans[i].DiskSpace,
				Authorized: *dt.Plans[i].Authorized,
			}
		}
		return c.outputFunc(&out, nil)
	}

	if c.ShowSettings != "" {
		var settings map[string]interface{}

		switch c.Name {
		case "grafana":
			if !utils.IsInList(grafanaSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(grafanaSettings, ", "),
				)
			}

			res, err := globalstate.EgoscaleClient.GetDbaasSettingsGrafanaWithResponse(ctx)
			if err != nil {
				return err
			}

			if c.ShowSettings == "grafana" {
				settings = *res.JSON200.Settings.Grafana.Properties
			}

			dbaasShowSettings(settings)
		case "kafka":
			if !utils.IsInList(kafkaSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(kafkaSettings, ", "),
				)
			}

			res, err := globalstate.EgoscaleClient.GetDbaasSettingsKafkaWithResponse(ctx)
			if err != nil {
				return err
			}

			switch c.ShowSettings {
			case "kafka":
				settings = *res.JSON200.Settings.Kafka.Properties
			case "kafka-connect":
				settings = *res.JSON200.Settings.KafkaConnect.Properties
			case "kafka-rest":
				settings = *res.JSON200.Settings.KafkaRest.Properties
			case "schema-registry":
				settings = *res.JSON200.Settings.SchemaRegistry.Properties
			}

			dbaasShowSettings(settings)

		case "opensearch":
			if !utils.IsInList(opensearchSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(opensearchSettings, ", "),
				)
			}

			res, err := globalstate.EgoscaleClient.GetDbaasSettingsOpensearchWithResponse(ctx)
			if err != nil {
				return err
			}

			if c.ShowSettings == "opensearch" {
				settings = *res.JSON200.Settings.Opensearch.Properties
			}

			dbaasShowSettings(settings)

		case "mysql":
			if !utils.IsInList(mysqlSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(mysqlSettings, ", "),
				)
			}

			res, err := globalstate.EgoscaleClient.GetDbaasSettingsMysqlWithResponse(ctx)
			if err != nil {
				return err
			}

			if c.ShowSettings == "mysql" {
				settings = *res.JSON200.Settings.Mysql.Properties
			}

			dbaasShowSettings(settings)

		case "pg":
			if !utils.IsInList(pgSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(pgSettings, ", "),
				)
			}

			res, err := globalstate.EgoscaleClient.GetDbaasSettingsPgWithResponse(ctx)
			if err != nil {
				return err
			}

			switch c.ShowSettings {
			case "pg":
				settings = *res.JSON200.Settings.Pg.Properties
			case "pgbouncer":
				settings = *res.JSON200.Settings.Pgbouncer.Properties
			case "pglookout":
				settings = *res.JSON200.Settings.Pglookout.Properties
			}

			dbaasShowSettings(settings)

		case "redis":
			if !utils.IsInList(redisSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(redisSettings, ", "),
				)
			}

			res, err := globalstate.EgoscaleClient.GetDbaasSettingsRedisWithResponse(ctx)
			if err != nil {
				return err
			}

			if c.ShowSettings == "redis" {
				settings = *res.JSON200.Settings.Redis.Properties
			}

			dbaasShowSettings(settings)
		}

		return nil
	}

	if c.ShowBackupConfig != "" {
		var bc *exo.DatabaseBackupConfig
		for _, plan := range dt.Plans {
			if *plan.Name == c.ShowBackupConfig {
				bc = plan.BackupConfig
				break
			}
		}
		if bc == nil {
			return fmt.Errorf("%q is not a valid plan", c.ShowBackupConfig)
		}
		return c.outputFunc(&dbaasTypePlanBackupOutput{
			Interval:                   bc.Interval,
			MaxCount:                   bc.MaxCount,
			RecoveryMode:               bc.RecoveryMode,
			FrequentIntervalMinutes:    bc.FrequentIntervalMinutes,
			FrequentOldestAgeMinutes:   bc.FrequentOldestAgeMinutes,
			InfrequentIntervalMinutes:  bc.InfrequentIntervalMinutes,
			InfrequentOldestAgeMinutes: bc.InfrequentOldestAgeMinutes,
		}, nil)
	}

	return c.outputFunc(&dbaasTypeShowOutput{
		Name:        *dt.Name,
		Description: utils.DefaultString(dt.Description, ""),
		AvailableVersions: func() (v []string) {
			if dt.AvailableVersions != nil {
				v = *dt.AvailableVersions
			}
			return
		}(),
		DefaultVersion: utils.DefaultString(dt.DefaultVersion, "-"),
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasTypeCmd, &dbaasTypeShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
