package cmd

import (
	"fmt"
	"strings"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbServiceUpdateCmd struct {
	_ bool `cli-cmd:"update"`

	Name string `cli-arg:"#"`

	MaintenanceDOW        string `cli-flag:"maintenance-dow" cli-usage:"automated Database Service maintenance day-of-week"`
	MaintenanceTime       string `cli-usage:"automated Database Service maintenance time (format HH:MM:SS)"`
	Plan                  string `cli-usage:"Database Service plan"`
	TerminationProtection bool   `cli-usage:"enable Database Service termination protection"`
	UserConfigFile        string `cli-flag:"user-config" cli-short:"c" cli-usage:"path to JSON user config file"`
	Zone                  string `cli-short:"z" cli-usage:"Database Service zone"`
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

func (c *dbServiceUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	databaseService, err := cs.GetDatabaseService(ctx, c.Zone, c.Name)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = c.Plan
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = c.TerminationProtection
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.UserConfigFile)) {
		if databaseService.UserConfig, err = getDatabaseServiceUserConfigFromFile(c.UserConfigFile); err != nil {
			return fmt.Errorf("error parsing user config: %s", err)
		}
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &exov2.DatabaseServiceMaintenance{
			DOW:  c.MaintenanceDOW,
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", databaseService.Name), func() {
		if updated {
			if err = cs.UpdateDatabaseService(ctx, c.Zone, databaseService); err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showDatabaseService(c.Zone, databaseService.Name))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbCmd, &dbServiceUpdateCmd{}))
}
