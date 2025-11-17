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

type DeploymentCreateCmd struct {
    exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name     string `cli-arg:"#" cli-usage:"NAME"`
	GPUType  string `cli-flag:"gpu-type" cli-usage:"GPU type family (e.g., gpua5000, gpu3080ti)"`
	GPUCount int64  `cli-flag:"gpu-count" cli-usage:"Number of GPUs (1-8)"`
	Replicas int64  `cli-flag:"replicas" cli-usage:"Number of replicas (>=1)"`

	ModelID   string      `cli-flag:"model-id" cli-usage:"Model ID (UUID)"`
	ModelName string      `cli-flag:"model-name" cli-usage:"Model name (as created)"`
	Zone      v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }
func (c *DeploymentCreateCmd) CmdShort() string     { return "Create AI deployment" }
func (c *DeploymentCreateCmd) CmdLong() string {
    return "This command creates an AI deployment on dedicated inference servers."
}
func (c *DeploymentCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
    exocmd.CmdSetZoneFlagFromDefault(cmd)
    return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *DeploymentCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if c.GPUType == "" || c.GPUCount == 0 {
		return fmt.Errorf("--gpu-type and --gpu-count are required")
	}
	if c.ModelID == "" && c.ModelName == "" {
		return fmt.Errorf("--model-id or --model-name is required")
	}

	req := v3.CreateDeploymentRequest{
		Name:     c.Name,
		GpuType:  c.GPUType,
		GpuCount: c.GPUCount,
		Replicas: c.Replicas,
	}
	if c.ModelID != "" || c.ModelName != "" {
		req.Model = &v3.ModelRef{}
		if c.ModelID != "" {
			if id, err := v3.ParseUUID(c.ModelID); err == nil {
				req.Model.ID = id
			} else {
				return fmt.Errorf("invalid --model-id: %w", err)
			}
		}
		if c.ModelName != "" {
			req.Model.Name = c.ModelName
		}
	}

	if err := utils.RunAsync(ctx, client, fmt.Sprintf("Creating deployment %q...", c.Name), func(ctx context.Context, c *v3.Client) (*v3.Operation, error) {
		return c.CreateDeployment(ctx, req)
	}); err != nil {
		return err
	}
	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Deployment created.")
	}
	return nil
}

func init() {
    cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
