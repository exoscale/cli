package dbaas

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasMigrationStopCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"stop"`

	Name string `cli-arg:"#"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasMigrationStopCmd) CmdAliases() []string { return []string{} }

func (c *dbaasMigrationStopCmd) CmdShort() string {
	return "Stop running migration of a Database"
}

func (c *dbaasMigrationStopCmd) CmdLong() string {
	return "This command stops the currently running migration of a Database."
}

func (c *dbaasMigrationStopCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasMigrationStopCmd) CmdRun(cmd *cobra.Command, args []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	var stopMigrationFuncs = map[v3.DBAASServiceTypeName]func(context.Context, string) (*v3.Operation, error){
		"mysql":  client.StopDBAASMysqlMigration,
		"pg":     client.StopDBAASPGMigration,
		"redis":  client.StopDBAASRedisMigration,
		"valkey": client.StopDBAASValkeyMigration,
	}

	if _, ok := stopMigrationFuncs[db.Type]; !ok {
		return fmt.Errorf("migrations not supported for database type %q", db.Type)
	}

	_, err = client.GetDBAASMigrationStatus(ctx, c.Name)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("migration for database %q not running in zone %q", c.Name, c.Zone)
		}
		return fmt.Errorf("failed to retrieve migration status: %s", err)
	}

	op, err := stopMigrationFuncs[db.Type](ctx, c.Name)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}
	utils.DecorateAsyncOperation("Stopping Database Migration...", func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return fmt.Errorf("failed to stop migration: %s", err)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasMigrationCmd, &dbaasMigrationStopCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
