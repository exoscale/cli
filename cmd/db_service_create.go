package cmd

import (
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbServiceCreateCmd struct {
	_ bool `cli-cmd:"create"`

	Type string `cli-arg:"#"`
	Plan string `cli-arg:"#"`
	Name string `cli-arg:"#"`

	MaintenanceDOW        string `cli-flag:"maintenance-dow" cli-usage:"automated Database Service maintenance day-of-week"`
	MaintenanceTime       string `cli-usage:"automated Database Service maintenance time (format HH:MM:SS)"`
	TerminationProtection bool   `cli-usage:"enable Database Service termination protection"`
	UserConfigFile        string `cli-flag:"user-config" cli-short:"c" cli-usage:"path to JSON user config file"`
	Zone                  string `cli-short:"z" cli-usage:"Database Service zone"`
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

func (c *dbServiceCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	databaseService := &egoscale.DatabaseService{
		Name:                  &c.Name,
		Plan:                  &c.Plan,
		TerminationProtection: &c.TerminationProtection,
		Type:                  &c.Type,
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &egoscale.DatabaseServiceMaintenance{
			DOW:  c.MaintenanceDOW,
			Time: c.MaintenanceTime,
		}
	}

	if c.UserConfigFile != "" {
		userConfig, err := getDatabaseServiceUserConfigFromFile(c.UserConfigFile)
		if err != nil {
			return fmt.Errorf("error parsing user config: %s", err)
		}
		databaseService.UserConfig = &userConfig
	}

	decorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", *databaseService.Name), func() {
		databaseService, err = cs.CreateDatabaseService(ctx, c.Zone, databaseService)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showDatabaseService(c.Zone, *databaseService.Name))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbCmd, &dbServiceCreateCmd{}))
}
