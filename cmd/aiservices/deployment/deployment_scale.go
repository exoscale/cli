package deployment

import (
	"context"
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type DeploymentScaleCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"scale"`

	Deployment string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Size       int64       `cli-arg:"#" cli-usage:"SIZE (replicas)"`
	Zone       v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentScaleCmd) CmdAliases() []string { return nil }
func (c *DeploymentScaleCmd) CmdShort() string     { return "Scale AI deployment" }
func (c *DeploymentScaleCmd) CmdLong() string {
	return "This command scales an AI deployment to the specified number of replicas."
}
func (c *DeploymentScaleCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *DeploymentScaleCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	// Resolve deployment ID using the SDK helper
	list, err := client.ListDeployments(ctx)
	if err != nil {
		return err
	}
	entry, err := list.FindListDeploymentsResponseEntry(c.Deployment)
	if err != nil {
		return err
	}
	id := entry.ID

	req := v3.ScaleDeploymentRequest{Replicas: c.Size}
	if err := utils.RunAsync(ctx, client, fmt.Sprintf("Scaling deployment %s to %d...", c.Deployment, c.Size), func(ctx context.Context, c *v3.Client) (*v3.Operation, error) {
		return c.ScaleDeployment(ctx, id, req)
	}); err != nil {
		return err
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentScaleCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
