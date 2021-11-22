package cmd

import (
	"fmt"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasServiceDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name string `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasServiceDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *dbaasServiceDeleteCmd) cmdShort() string { return "Delete a Database Service" }

func (c *dbaasServiceDeleteCmd) cmdLong() string { return "" }

func (c *dbaasServiceDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Database Service %q?", c.Name)) {
			return nil
		}
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Deleting Database Service %q...", c.Name), func() {
		err = cs.DeleteDatabaseService(ctx, c.Zone, &egoscale.DatabaseService{Name: &c.Name})
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasCmd, &dbaasServiceDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
