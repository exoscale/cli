package cmd

import (
	"errors"
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasMigrationStatusCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"status"`

	Name string `cli-arg:"#"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasMigrationStatusCmd) cmdAliases() []string { return []string{} }

func (c *dbaasMigrationStatusCmd) cmdShort() string {
	return "Migration status of a Database"
}

func (c *dbaasMigrationStatusCmd) cmdLong() string {
	return "This command shows the status of the migration of a Database."
}

func (c *dbaasMigrationStatusCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

type databaseMigrationStatus v2.DatabaseMigrationStatus

func (o *databaseMigrationStatus) toJSON()  { output.JSON(o) }
func (o *databaseMigrationStatus) toText()  { output.Text(o) }
func (o *databaseMigrationStatus) toTable() { output.Table(o) }

func (c *dbaasMigrationStatusCmd) cmdRun(cmd *cobra.Command, args []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	res, err := globalstate.GlobalEgoscaleClient.GetDatabaseMigrationStatus(ctx, c.Zone, c.Name)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return fmt.Errorf("failed to retrieve migration status: %s", err)
	}

	return c.outputFunc((*databaseMigrationStatus)(res), nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasMigrationCmd, &dbaasMigrationStatusCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
