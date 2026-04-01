package deployment

import (
	"context"
	"fmt"
	"os"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type DeploymentUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Deployment string `cli-arg:"#" cli-usage:"ID or NAME"`

	Name                      string      `cli-flag:"name" cli-usage:"New deployment name"`
	InferenceEngineParameters string      `cli-flag:"inference-engine-params" cli-usage:"Space-separated inference engine server CLI arguments (e.g., \"--gpu-memory-usage=0.8 --max-tokens=4096\")"`
	InferenceEngineVersion    string      `cli-flag:"inference-engine-version" cli-usage:"Inference engine version"`
	InferenceEngineHelp       bool        `cli-flag:"inference-engine-parameter-help" cli-usage:"Show inference engine parameters help"`
	Zone                      v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentUpdateCmd) CmdAliases() []string { return nil }
func (c *DeploymentUpdateCmd) CmdShort() string     { return "Update AI deployment" }
func (c *DeploymentUpdateCmd) CmdLong() string {
	return "This command updates an AI deployment."
}
func (c *DeploymentUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *DeploymentUpdateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if c.InferenceEngineHelp {
		// Reusing the help logic from DeploymentCreateCmd if possible, but since it's a method on DeploymentCreateCmd
		// and we are in the same package, we might want to make it a package-level utility or just instantiate DeploymentCreateCmd.
		createCmd := &DeploymentCreateCmd{}
		return createCmd.showInferenceEngineParameterHelp(ctx, client, c.InferenceEngineVersion)
	}

	if c.Deployment == "" {
		return fmt.Errorf("deployment ID or name is required")
	}

	// Resolve deployment ID
	list, err := client.ListDeployments(ctx)
	if err != nil {
		return err
	}
	entry, err := list.FindListDeploymentsResponseEntry(c.Deployment)
	if err != nil {
		return err
	}
	id := entry.ID

	req := v3.UpdateDeploymentRequest{
		Name:                   c.Name,
		InferenceEngineVersion: v3.InferenceEngineVersion(c.InferenceEngineVersion),
	}

	if c.InferenceEngineParameters != "" {
		req.InferenceEngineParameters = strings.Fields(c.InferenceEngineParameters)
	}

	if err := utils.RunAsync(ctx, client, fmt.Sprintf("Updating deployment %q...", c.Deployment), func(ctx context.Context, c *v3.Client) (*v3.Operation, error) {
		return c.UpdateDeployment(ctx, id, req)
	}); err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Deployment updated.")
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentUpdateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
