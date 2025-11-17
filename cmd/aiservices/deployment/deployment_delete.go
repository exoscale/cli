package deployment

import (
	"context"
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type DeploymentDeleteCmd struct {
    exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Deployment string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Zone       v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentDeleteCmd) CmdAliases() []string { return exocmd.GDeleteAlias }
func (c *DeploymentDeleteCmd) CmdShort() string     { return "Delete AI deployment" }
func (c *DeploymentDeleteCmd) CmdLong() string {
    return "This command deletes an AI deployment by ID or name."
}
func (c *DeploymentDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
    exocmd.CmdSetZoneFlagFromDefault(cmd)
    return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *DeploymentDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := ResolveDeploymentID(ctx, client, c.Deployment)
	if err != nil {
		return err
	}

	if err := utils.RunAsync(ctx, client, fmt.Sprintf("Deleting deployment %s...", c.Deployment), func(ctx context.Context, c *v3.Client) (*v3.Operation, error) {
		return c.DeleteDeployment(ctx, id)
	}); err != nil {
		return err
	}
	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Deployment deleted.")
	}
	return nil
}

func init() {
    cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
