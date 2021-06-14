package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbServiceDeleteCmd struct {
	_ bool `cli-cmd:"delete"`

	Name string `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbServiceDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *dbServiceDeleteCmd) cmdShort() string { return "Delete a Database Service" }

func (c *dbServiceDeleteCmd) cmdLong() string { return "" }

func (c *dbServiceDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbServiceDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Database Service %q?", c.Name)) {
			return nil
		}
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Deleting Database Service %q...", c.Name), func() {
		err = cs.DeleteDatabaseService(ctx, c.Zone, c.Name)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbCmd, &dbServiceDeleteCmd{}))
}
