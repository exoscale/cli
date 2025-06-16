package dbaas

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasMigrationStatusCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"status"`

	Name string `cli-arg:"#"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasMigrationStatusCmd) CmdAliases() []string { return []string{} }

func (c *dbaasMigrationStatusCmd) CmdShort() string {
	return "Migration status of a Database"
}

func (c *dbaasMigrationStatusCmd) CmdLong() string {
	return "This command shows the status of the migration of a Database."
}

func (c *dbaasMigrationStatusCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

type databaseMigrationStatus v3.DBAASMigrationStatus

func (o *databaseMigrationStatus) ToJSON()  { output.JSON(o) }
func (o *databaseMigrationStatus) ToText()  { output.Text(o) }
func (o *databaseMigrationStatus) ToTable() { output.Table(o) }

func (c *dbaasMigrationStatusCmd) CmdRun(cmd *cobra.Command, args []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	res, err := client.GetDBAASMigrationStatus(ctx, c.Name)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return fmt.Errorf("failed to retrieve migration status: %s", err)
	}

	return c.OutputFunc((*databaseMigrationStatus)(res), nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasMigrationCmd, &dbaasMigrationStatusCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
