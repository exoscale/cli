package dbaas

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
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

	t.Append([]string{"Backup interval (hours)", utils.Int64PtrFormatOutput(o.Interval)})
	t.Append([]string{"Max backups", utils.Int64PtrFormatOutput(o.MaxCount)})
	t.Append([]string{"Recovery mode", utils.DefaultString(o.RecoveryMode, "")})
	t.Append([]string{"Frequent backup interval", utils.Int64PtrFormatOutput(o.FrequentIntervalMinutes)})
	t.Append([]string{"Frequent backup max age", utils.Int64PtrFormatOutput(o.FrequentOldestAgeMinutes)})
	t.Append([]string{"Infrequent backup interval", utils.Int64PtrFormatOutput(o.InfrequentIntervalMinutes)})
	t.Append([]string{"Infrequent backup max age", utils.Int64PtrFormatOutput(o.InfrequentOldestAgeMinutes)})
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
	redisSettings  = []string{"redis"}
	valkeySettings = []string{"valkey"}
)

type dbaasTypeShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string `cli-arg:"#"`

	ShowPlans        bool   `cli-flag:"plans" cli-usage:"list plans offered for the Database Service type"`
	ShowSettings     string `cli-flag:"settings" cli-usage:"show settings supported by the Database Service type"`
	ShowBackupConfig string `cli-flag:"backup-config" cli-usage:"show backup configuration for the Database Service type and Plan"`
}

func (c *dbaasTypeShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *dbaasTypeShowCmd) CmdShort() string { return "Show a Database Service type details" }

func (c *dbaasTypeShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Database Service type details.

Supported Database Service type settings:

* %s
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
		strings.Join(valkeySettings, ", "),
		strings.Join(output.TemplateAnnotations(&dbaasTypeShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasTypePlanListItemOutput{}), ", "))
}

func (c *dbaasTypeShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasTypeShowCmd) CmdRun(_ *cobra.Command, _ []string) error { //nolint:gocyclo
	ctx := exocmd.GContext
	var err error

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	dt, err := client.GetDBAASServiceType(ctx, c.Name)
	if err != nil {
		return err
	}

	if c.ShowPlans {
		out := make(dbaasTypePlanListOutput, len(dt.Plans))
		for i := range dt.Plans {
			out[i] = dbaasTypePlanListItemOutput{
				Name:       dt.Plans[i].Name,
				Nodes:      dt.Plans[i].NodeCount,
				NodeCPUs:   dt.Plans[i].NodeCPUCount,
				NodeMemory: dt.Plans[i].NodeMemory,
				DiskSpace:  dt.Plans[i].NodeMemory,
				Authorized: utils.DefaultBool(dt.Plans[i].Authorized, false),
			}
		}
		return c.OutputFunc(&out, nil)
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

			res, err := client.GetDBAASSettingsGrafana(ctx)
			if err != nil {
				return err
			}

			if c.ShowSettings == "grafana" {
				settings = res.Settings.Grafana.Properties
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

			res, err := client.GetDBAASSettingsKafka(ctx)
			if err != nil {
				return err
			}

			switch c.ShowSettings {
			case "kafka":
				settings = res.Settings.Kafka.Properties
			case "kafka-connect":
				settings = res.Settings.KafkaConnect.Properties
			case "kafka-rest":
				settings = res.Settings.KafkaRest.Properties
			case "schema-registry":
				settings = res.Settings.SchemaRegistry.Properties
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

			res, err := client.GetDBAASSettingsOpensearch(ctx)
			if err != nil {
				return err
			}

			if c.ShowSettings == "opensearch" {
				settings = res.Settings.Opensearch.Properties
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

			res, err := client.GetDBAASSettingsMysql(ctx)
			if err != nil {
				return err
			}

			if c.ShowSettings == "mysql" {
				settings = res.Settings.Mysql.Properties
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

			res, err := client.GetDBAASSettingsPG(ctx)
			if err != nil {
				return err
			}

			switch c.ShowSettings {
			case "pg":
				settings = res.Settings.PG.Properties
			case "pgbouncer":
				settings = res.Settings.Pgbouncer.Properties
			case "pglookout":
				settings = res.Settings.Pglookout.Properties
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

			res, err := client.GetDBAASSettingsRedis(ctx)
			if err != nil {
				return err
			}

			if c.ShowSettings == "redis" {
				settings = res.Settings.Redis.Properties
			}

			dbaasShowSettings(settings)

		case "valkey":
			if !utils.IsInList(valkeySettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(valkeySettings, ", "),
				)
			}

			res, err := client.GetDBAASSettingsValkey(ctx)

			if err != nil {
				return err
			}

			if c.ShowSettings == "valkey" {
				settings = res.Settings.Valkey.Properties
			}

			dbaasShowSettings(settings)
		}

		return nil
	}

	if c.ShowBackupConfig != "" {
		var bc *v3.DBAASBackupConfig
		for _, plan := range dt.Plans {
			if plan.Name == c.ShowBackupConfig {
				bc = plan.BackupConfig
				break
			}
		}
		if bc == nil {
			return fmt.Errorf("%q is not a valid plan", c.ShowBackupConfig)
		}
		return c.OutputFunc(&dbaasTypePlanBackupOutput{
			Interval:                   &bc.Interval,
			MaxCount:                   &bc.MaxCount,
			RecoveryMode:               &bc.RecoveryMode,
			FrequentIntervalMinutes:    &bc.FrequentIntervalMinutes,
			FrequentOldestAgeMinutes:   &bc.FrequentOldestAgeMinutes,
			InfrequentIntervalMinutes:  &bc.InfrequentIntervalMinutes,
			InfrequentOldestAgeMinutes: &bc.InfrequentOldestAgeMinutes,
		}, nil)
	}

	return c.OutputFunc(&dbaasTypeShowOutput{
		Name:        string(dt.Name),
		Description: dt.Description,
		AvailableVersions: func() (v []string) {
			if dt.AvailableVersions != nil {
				v = dt.AvailableVersions
			}
			return
		}(),
		DefaultVersion: utils.DefaultString(&dt.DefaultVersion, "-"),
	}, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasTypeCmd, &dbaasTypeShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
