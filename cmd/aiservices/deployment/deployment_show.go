package deployment

import (
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type DeploymentShowOutput struct {
	ID            v3.UUID                        `json:"id"`
	Name          string                         `json:"name"`
	Status        v3.GetDeploymentResponseStatus `json:"status"`
	StatusDetails string                         `json:"status_details"`
	GPUType       string                         `json:"gpu_type"`
	GPUCount      int64                          `json:"gpu_count"`
	Replicas      int64                          `json:"replicas"`
	ServiceLevel  string                         `json:"service_level"`
	DeploymentURL string                         `json:"deployment_url"`
	ModelID       v3.UUID                        `json:"model_id"`
	ModelName     string                         `json:"model_name"`
	CreatedAt     string                         `json:"created_at"`
	UpdatedAt     string                         `json:"updated_at"`
}

func (o *DeploymentShowOutput) ToJSON()  { output.JSON(o) }
func (o *DeploymentShowOutput) ToText()  { output.Text(o) }
func (o *DeploymentShowOutput) ToTable() { output.Table(o) }

type DeploymentShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Deployment string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Zone       v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentShowCmd) CmdAliases() []string { return exocmd.GShowAlias }
func (c *DeploymentShowCmd) CmdShort() string     { return "Show AI deployment" }
func (c *DeploymentShowCmd) CmdLong() string {
	return "This command shows details of an AI deployment by ID or name."
}
func (c *DeploymentShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *DeploymentShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	resp, err := client.GetDeployment(ctx, id)
	if err != nil {
		return err
	}

	var modelID v3.UUID
	var modelName string
	if resp.Model != nil {
		modelID = resp.Model.ID
		modelName = resp.Model.Name
	}

	out := &DeploymentShowOutput{
		ID:            resp.ID,
		Name:          resp.Name,
		Status:        resp.Status,
		StatusDetails: resp.StatusDetails,
		GPUType:       resp.GpuType,
		GPUCount:      resp.GpuCount,
		Replicas:      resp.Replicas,
		ServiceLevel:  resp.ServiceLevel,
		DeploymentURL: resp.DeploymentURL,
		ModelID:       modelID,
		ModelName:     modelName,
		CreatedAt:     resp.CreatedAT.Format(time.RFC3339),
		UpdatedAt:     resp.UpdatedAT.Format(time.RFC3339),
	}
	return c.OutputFunc(out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
