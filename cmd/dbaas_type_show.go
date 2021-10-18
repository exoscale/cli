package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"
)

type dbTypeShowOutput struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	LatestVersion  string   `json:"latest_version"`
	DefaultVersion string   `json:"default_version"`
	Plans          []string `json:"plans"`
}

func (o *dbTypeShowOutput) toJSON() { outputJSON(o) }
func (o *dbTypeShowOutput) toText() { outputText(o) }
func (o *dbTypeShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Latest Version", o.LatestVersion})
	t.Append([]string{"Default Version", o.DefaultVersion})
	t.Append([]string{"Plans", strings.Join(o.Plans, "\n")})
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

type dbTypeShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string `cli-arg:"#"`

	ShowSettings string `cli-flag:"settings" cli-usage:"show supported settings by the Database Service type"`
}

func (c *dbTypeShowCmd) cmdAliases() []string { return gShowAlias }

func (c *dbTypeShowCmd) cmdShort() string { return "Show a Database Service type details" }

func (c *dbTypeShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Database Service type details.

Supported Database Service type settings:

* %s
* %s
* %s
* %s

Supported output template annotations: %s

Note: plans marked with (U) are currently unauthorized for this organization.`,
		strings.Join(kafkaSettings, ", "),
		strings.Join(mysqlSettings, ", "),
		strings.Join(pgSettings, ", "),
		strings.Join(redisSettings, ", "),
		strings.Join(outputterTemplateAnnotations(&dbTypeShowOutput{}), ", "))
}

func (c *dbTypeShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbTypeShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	dt, err := cs.GetDatabaseServiceType(ctx, gCurrentAccount.DefaultZone, c.Name)
	if err != nil {
		return err
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

			c.showSettings(settings)

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

			c.showSettings(settings)

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

			c.showSettings(settings)

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

			c.showSettings(settings)
		}

		return nil
	}

	return output(&dbTypeShowOutput{
		Name:           *dt.Name,
		Description:    defaultString(dt.Description, ""),
		LatestVersion:  defaultString(dt.LatestVersion, "-"),
		DefaultVersion: defaultString(dt.DefaultVersion, "-"),
		Plans: func() []string {
			plans := make([]string, len(dt.Plans))
			for i := range dt.Plans {
				plans[i] = *dt.Plans[i].Name
				if !defaultBool(dt.Plans[i].Authorized, false) {
					plans[i] += " (U)"
				}
			}
			return plans
		}(),
	}, nil)
}

func (c *dbTypeShowCmd) showSettings(settings map[string]interface{}) {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{"key", "type", "description"})

	for k, v := range settings {
		s, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		row := []string{k}

		typ := "-"
		if v, ok := s["type"]; ok {
			typ = fmt.Sprint(v)
		}
		row = append(row, typ)

		var description string
		if v, ok := s["description"]; ok {
			description = wordwrap.WrapString(v.(string), 50)

			if v, ok := s["enum"]; ok {
				description = fmt.Sprintf("%s\n  * Supported values:\n%s", description, func() string {
					values := make([]string, len(v.([]interface{})))
					for i, val := range v.([]interface{}) {
						values[i] = fmt.Sprintf("    - %v", val)
					}
					return strings.Join(values, "\n")
				}())
			}

			min, hasMin := s["minimum"]
			max, hasMax := s["maximum"]
			if hasMin && hasMax {
				description = fmt.Sprintf("%s\n  * Minimum: %v / Maximum: %v", description, min, max)
			}

			if v, ok := s["default"]; ok {
				description = fmt.Sprintf("%s\n  * Default: %v", description, v)
			}

			if v, ok := s["example"]; ok {
				description = fmt.Sprintf("%s\n  * Example: %v", description, v)
			}
		}
		row = append(row, description)

		t.Append(row)
	}
}

func init() {
	cobra.CheckErr(registerCLICommand(dbTypeCmd, &dbTypeShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
