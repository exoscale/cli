package deployment

import (
	"fmt"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type DeploymentListItemOutput struct {
	ID        v3.UUID                               `json:"id"`
	Name      string                                `json:"name"`
	Status    v3.ListDeploymentsResponseEntryStatus `json:"status"`
	GPUType   string                                `json:"gpu_type"`
	GPUCount  int64                                 `json:"gpu_count"`
	Replicas  int64                                 `json:"replicas"`
	ModelName string                                `json:"model_name"`
}

type DeploymentListOutput []DeploymentListItemOutput

func (o *DeploymentListOutput) ToJSON()  { output.JSON(o) }
func (o *DeploymentListOutput) ToText()  { output.Text(o) }
func (o *DeploymentListOutput) ToTable() { output.Table(o) }

type DeploymentListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *DeploymentListCmd) CmdShort() string     { return "List AI deployments" }
func (c *DeploymentListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists AI deployments.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&DeploymentListOutput{}), ", "))
}
func (c *DeploymentListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *DeploymentListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListDeployments(ctx)
	if err != nil {
		return err
	}

	out := make(DeploymentListOutput, 0, len(resp.Deployments))
	for _, d := range resp.Deployments {
		var modelName string
		if d.Model != nil {
			modelName = d.Model.Name
		}
		out = append(out, DeploymentListItemOutput{
			ID:        d.ID,
			Name:      d.Name,
			Status:    d.Status,
			GPUType:   d.GpuType,
			GPUCount:  d.GpuCount,
			Replicas:  d.Replicas,
			ModelName: modelName,
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
    cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
