package deployment

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type DeploymentLogsCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"logs"`

	Deployment string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Zone       v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentLogsCmd) CmdAliases() []string { return nil }
func (c *DeploymentLogsCmd) CmdShort() string     { return "Get deployment logs" }
func (c *DeploymentLogsCmd) CmdLong() string {
	return "This command retrieves logs for the deployment's vLLM component."
}
func (c *DeploymentLogsCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *DeploymentLogsCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := ResolveDeploymentID(ctx, client, c.Deployment)
	if err != nil {
		return err
	}

	resp, err := client.GetDeploymentLogs(ctx, id)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, string(*resp))
		return nil
	}
	// When quiet, do nothing
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentLogsCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
