package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
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

func (o *dbaasTypePlanListOutput) toJSON() { output.JSON(o) }
func (o *dbaasTypePlanListOutput) toText() { output.Text(o) }
func (o *dbaasTypePlanListOutput) toTable() {
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

func (o *dbaasTypePlanBackupOutput) toJSON() { output.JSON(o) }
func (o *dbaasTypePlanBackupOutput) toText() { output.Text(o) }
func (o *dbaasTypePlanBackupOutput) toTable() {
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

func (o *dbaasTypeShowOutput) toJSON() { output.JSON(o) }
func (o *dbaasTypeShowOutput) toText() { output.Text(o) }
func (o *dbaasTypeShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Available Versions", strings.Join(o.AvailableVersions, ", ")})
	t.Append([]string{"Default Version", o.DefaultVersion})
}

var (
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

Supported output template annotations:

* When showing a Database Service: %s

* When listing Database Service plans: %s`,
		strings.Join(opensearchSettings, ", "),
		strings.Join(kafkaSettings, ", "),
		strings.Join(mysqlSettings, ", "),
		strings.Join(pgSettings, ", "),
		strings.Join(redisSettings, ", "),
		strings.Join(output.OutputterTemplateAnnotations(&dbaasTypeShowOutput{}), ", "),
		strings.Join(output.OutputterTemplateAnnotations(&dbaasTypePlanListItemOutput{}), ", "))
}

func (c *dbaasTypeShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasTypeShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	dt, err := globalstate.GlobalEgoscaleClient.GetDatabaseServiceType(ctx, gCurrentAccount.DefaultZone, c.Name)
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
		case "kafka":
			if !utils.IsInList(kafkaSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(kafkaSettings, ", "),
				)
			}

			res, err := globalstate.GlobalEgoscaleClient.GetDbaasSettingsKafkaWithResponse(ctx)
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

			res, err := globalstate.GlobalEgoscaleClient.GetDbaasSettingsOpensearchWithResponse(ctx)
			if err != nil {
				return err
			}

			switch c.ShowSettings {
			case "opensearch":
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

			res, err := globalstate.GlobalEgoscaleClient.GetDbaasSettingsMysqlWithResponse(ctx)
			if err != nil {
				return err
			}

			switch c.ShowSettings {
			case "mysql":
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

			res, err := globalstate.GlobalEgoscaleClient.GetDbaasSettingsPgWithResponse(ctx)
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

			res, err := globalstate.GlobalEgoscaleClient.GetDbaasSettingsRedisWithResponse(ctx)
			if err != nil {
				return err
			}

			switch c.ShowSettings {
			case "redis":
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
