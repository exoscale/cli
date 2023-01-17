package cmd

import (
	"errors"
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasMigrationStopCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"stop"`

	Name string `cli-arg:"#"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasMigrationStopCmd) cmdAliases() []string { return []string{} }

func (c *dbaasMigrationStopCmd) cmdShort() string {
	return "Stop running migration of a Database"
}

func (c *dbaasMigrationStopCmd) cmdLong() string {
	return "This command stops the currently running migration of a Database."
}

func (c *dbaasMigrationStopCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasMigrationStopCmd) cmdRun(cmd *cobra.Command, args []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	dbType, err := dbaasGetType(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	decorateAsyncOperation("Stopping Database Migration...", func() {
		switch dbType {
		case "mysql":
			err = cs.StopMysqlDatabaseMigration(ctx, c.Zone, c.Name)
		case "pg":
			err = cs.StopPgDatabaseMigration(ctx, c.Zone, c.Name)
		case "redis":
			err = cs.StopRedisDatabaseMigration(ctx, c.Zone, c.Name)
		default:
			err = fmt.Errorf("migrations not supported for database type %q", dbType)
		}
	})

	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("migration not running in zone %q", c.Zone)
		}
		return fmt.Errorf("failed to stop migration: %s", err)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasMigrationCmd, &dbaasMigrationStopCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
