package deployment

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type DeploymentRevealAPIKeyOutput struct {
    APIKey string `json:"api_key"`
}

func (o *DeploymentRevealAPIKeyOutput) ToJSON()  { output.JSON(o) }
func (o *DeploymentRevealAPIKeyOutput) ToText()  { output.Text(o) }
func (o *DeploymentRevealAPIKeyOutput) ToTable() { output.Table(o) }

type DeploymentRevealAPIKeyCmd struct {
    exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reveal-api-key"`

	Deployment string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Zone       v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentRevealAPIKeyCmd) CmdAliases() []string { return nil }
func (c *DeploymentRevealAPIKeyCmd) CmdShort() string     { return "Reveal deployment API key" }
func (c *DeploymentRevealAPIKeyCmd) CmdLong() string {
    return "This command reveals the inference endpoint API key for the deployment."
}
func (c *DeploymentRevealAPIKeyCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
    exocmd.CmdSetZoneFlagFromDefault(cmd)
    return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *DeploymentRevealAPIKeyCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := ResolveDeploymentID(ctx, client, c.Deployment)
	if err != nil {
		return err
	}

	resp, err := client.RevealDeploymentAPIKey(ctx, id)
	if err != nil {
		return err
	}

    out := &DeploymentRevealAPIKeyOutput{APIKey: resp.APIKey}
    return c.OutputFunc(out, nil)
}

func init() {
    cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentRevealAPIKeyCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
