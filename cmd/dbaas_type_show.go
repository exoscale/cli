package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/table"
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

func (o *dbaasTypePlanListOutput) toJSON() { outputJSON(o) }
func (o *dbaasTypePlanListOutput) toText() { outputText(o) }
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

type dbaasTypeShowOutput struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	AvailableVersions []string `json:"available_versions"`
	DefaultVersion    string   `json:"default_version"`
}

func (o *dbaasTypeShowOutput) toJSON() { outputJSON(o) }
func (o *dbaasTypeShowOutput) toText() { outputText(o) }
func (o *dbaasTypeShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Available Versions", strings.Join(o.AvailableVersions, ", ")})
	t.Append([]string{"Default Version", o.DefaultVersion})
}

var (
	kafkaSettings = []string{
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

	ShowPlans    bool   `cli-flag:"plans" cli-usage:"list plans offered for the Database Service type"`
	ShowSettings string `cli-flag:"settings" cli-usage:"show settings supported by the Database Service type"`
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

Supported output template annotations:

* When showing a Database Service: %s

* When listing Database Service plans: %s`,
		strings.Join(kafkaSettings, ", "),
		strings.Join(mysqlSettings, ", "),
		strings.Join(pgSettings, ", "),
		strings.Join(redisSettings, ", "),
		strings.Join(outputterTemplateAnnotations(&dbaasTypeShowOutput{}), ", "),
		strings.Join(outputterTemplateAnnotations(&dbaasTypePlanListItemOutput{}), ", "))
}

func (c *dbaasTypeShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasTypeShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	dt, err := cs.GetDatabaseServiceType(ctx, gCurrentAccount.DefaultZone, c.Name)
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
			if !isInList(kafkaSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(kafkaSettings, ", "),
				)
			}

			res, err := cs.GetDbaasSettingsKafkaWithResponse(ctx)
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

		case "mysql":
			if !isInList(mysqlSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(mysqlSettings, ", "),
				)
			}

			res, err := cs.GetDbaasSettingsMysqlWithResponse(ctx)
			if err != nil {
				return err
			}

			switch c.ShowSettings {
			case "mysql":
				settings = *res.JSON200.Settings.Mysql.Properties
			}

			dbaasShowSettings(settings)

		case "pg":
			if !isInList(pgSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(pgSettings, ", "),
				)
			}

			res, err := cs.GetDbaasSettingsPgWithResponse(ctx)
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
			if !isInList(redisSettings, c.ShowSettings) {
				return fmt.Errorf(
					"invalid settings value %q, expected one of: %s",
					c.ShowSettings,
					strings.Join(redisSettings, ", "),
				)
			}

			res, err := cs.GetDbaasSettingsRedisWithResponse(ctx)
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

	return c.outputFunc(&dbaasTypeShowOutput{
		Name:        *dt.Name,
		Description: defaultString(dt.Description, ""),
		AvailableVersions: func() (v []string) {
			if dt.AvailableVersions != nil {
				v = *dt.AvailableVersions
			}
			return
		}(),
		DefaultVersion: defaultString(dt.DefaultVersion, "-"),
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasTypeCmd, &dbaasTypeShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
