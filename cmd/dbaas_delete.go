package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasServiceDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name string `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasServiceDeleteCmd) CmdAliases() []string { return GRemoveAlias }

func (c *dbaasServiceDeleteCmd) CmdShort() string { return "Delete a Database Service" }

func (c *dbaasServiceDeleteCmd) CmdLong() string { return "" }

func (c *dbaasServiceDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	var err error

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Database Service %q?", c.Name)) {
			return nil
		}
	}

	op, err := client.DeleteDBAASService(ctx, c.Name)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Database Service %q...", c.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(dbaasCmd, &dbaasServiceDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
