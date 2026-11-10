package vpc

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcDeleteCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	VPC string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcDeleteCmd) CmdAliases() []string { return clicmd.GRemoveAlias }

func (c *vpcDeleteCmd) CmdShort() string { return "Delete a VPC" }

func (c *vpcDeleteCmd) CmdLong() string { return "" }

func (c *vpcDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	clicmd.CmdSetZoneFlagFromDefault(cmd)
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	entry, err := FindVPC(ctx, client, c.VPC)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete VPC %s?", c.VPC)) {
			return nil
		}
	}

	if err := client.DeleteVpc(ctx, entry.ID); err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &vpcDeleteCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
