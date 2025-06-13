package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type dbaasServiceDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name string `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasServiceDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *dbaasServiceDeleteCmd) CmdShort() string { return "Delete a Database Service" }

func (c *dbaasServiceDeleteCmd) CmdLong() string { return "" }

func (c *dbaasServiceDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	if !c.Force {
		if !utils.AskQuestion(fmt.Sprintf("Are you sure you want to delete Database Service %q?", c.Name)) {
			return nil
		}
	}

	var err error
	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Database Service %q...", c.Name), func() {
		err = globalstate.EgoscaleClient.DeleteDatabaseService(ctx, c.Zone, &egoscale.DatabaseService{Name: &c.Name})
	})
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasCmd, &dbaasServiceDeleteCmd{
		cliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
