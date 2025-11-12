package dedicated_inference

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

var deploymentCmd = &cobra.Command{
	Use:   "deployment",
	Short: "Manage AI deployments",
}

func init() {
	DedicatedInferenceCmd.AddCommand(deploymentCmd)
}

// list

type deploymentListItemOutput struct {
	ID        v3.UUID                               `json:"id"`
	Name      string                                `json:"name"`
	Status    v3.ListDeploymentsResponseEntryStatus `json:"status"`
	GPUType   string                                `json:"gpu_type"`
	GPUCount  int64                                 `json:"gpu_count"`
	Replicas  int64                                 `json:"replicas"`
	ModelName string                                `json:"model_name"`
}

type deploymentListOutput []deploymentListItemOutput

func (o *deploymentListOutput) ToJSON()  { output.JSON(o) }
func (o *deploymentListOutput) ToText()  { output.Text(o) }
func (o *deploymentListOutput) ToTable() { output.Table(o) }

type deploymentListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *deploymentListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *deploymentListCmd) CmdShort() string     { return "List AI deployments" }
func (c *deploymentListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists AI deployments.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&deploymentListOutput{}), ", "))
}
func (c *deploymentListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *deploymentListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListDeployments(ctx)
	if err != nil {
		return err
	}

	out := make(deploymentListOutput, 0, len(resp.Deployments))
	for _, d := range resp.Deployments {
		var modelName string
		if d.Model != nil {
			modelName = d.Model.Name
		}
		out = append(out, deploymentListItemOutput{
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

// create

type deploymentCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name     string `cli-flag:"name" cli-usage:"Deployment name"`
	GPUType  string `cli-flag:"gpu-type" cli-usage:"GPU type family (e.g., gpua5000, gpu3080ti)"`
	GPUCount int64  `cli-flag:"gpu-count" cli-usage:"Number of GPUs (1-8)"`
	Replicas int64  `cli-flag:"replicas" cli-usage:"Number of replicas (>=1)"`

	ModelID   string      `cli-flag:"model-id" cli-usage:"Model ID (UUID)"`
	ModelName string      `cli-flag:"model-name" cli-usage:"Model name (as created)"`
	Zone      v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *deploymentCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }
func (c *deploymentCreateCmd) CmdShort() string     { return "Create AI deployment" }
func (c *deploymentCreateCmd) CmdLong() string {
	return "This command creates an AI deployment on dedicated inference servers."
}
func (c *deploymentCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *deploymentCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	if err := runAsync(ctx, client, fmt.Sprintf("Creating deployment %q...", c.Name), func(ctx context.Context, c *v3.Client) (*v3.Operation, error) {
		return c.CreateDeployment(ctx, req)
	}); err != nil {
		return err
	}
	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Deployment created.")
	}
	return nil
}

// delete

type deploymentDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Deployment string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Zone       v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *deploymentDeleteCmd) CmdAliases() []string { return exocmd.GDeleteAlias }
func (c *deploymentDeleteCmd) CmdShort() string     { return "Delete AI deployment" }
func (c *deploymentDeleteCmd) CmdLong() string {
	return "This command deletes an AI deployment by ID or name."
}
func (c *deploymentDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *deploymentDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := resolveDeploymentID(ctx, client, c.Deployment)
	if err != nil {
		return err
	}

	if err := runAsync(ctx, client, fmt.Sprintf("Deleting deployment %s...", c.Deployment), func(ctx context.Context, c *v3.Client) (*v3.Operation, error) {
		return c.DeleteDeployment(ctx, id)
	}); err != nil {
		return err
	}
	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Deployment deleted.")
	}
	return nil
}

// scale

type deploymentScaleCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"scale"`

	Deployment string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Size       int64       `cli-arg:"#" cli-usage:"SIZE (replicas)"`
	Zone       v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *deploymentScaleCmd) CmdAliases() []string { return nil }
func (c *deploymentScaleCmd) CmdShort() string     { return "Scale AI deployment" }
func (c *deploymentScaleCmd) CmdLong() string {
	return "This command scales an AI deployment to the specified number of replicas."
}
func (c *deploymentScaleCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *deploymentScaleCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := resolveDeploymentID(ctx, client, c.Deployment)
	if err != nil {
		return err
	}

	req := v3.ScaleDeploymentRequest{Replicas: c.Size}
	if err := runAsync(ctx, client, fmt.Sprintf("Scaling deployment %s to %d...", c.Deployment, c.Size), func(ctx context.Context, c *v3.Client) (*v3.Operation, error) {
		return c.ScaleDeployment(ctx, id, req)
	}); err != nil {
		return err
	}
	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Deployment scaled.")
	}
	return nil
}

// reveal api key

type deploymentRevealAPIKeyOutput struct {
	APIKey string `json:"api_key"`
}

func (o *deploymentRevealAPIKeyOutput) ToJSON()  { output.JSON(o) }
func (o *deploymentRevealAPIKeyOutput) ToText()  { output.Text(o) }
func (o *deploymentRevealAPIKeyOutput) ToTable() { output.Table(o) }

type deploymentRevealAPIKeyCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reveal-api-key"`

	Deployment string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Zone       v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *deploymentRevealAPIKeyCmd) CmdAliases() []string { return nil }
func (c *deploymentRevealAPIKeyCmd) CmdShort() string     { return "Reveal deployment API key" }
func (c *deploymentRevealAPIKeyCmd) CmdLong() string {
	return "This command reveals the inference endpoint API key for the deployment."
}
func (c *deploymentRevealAPIKeyCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *deploymentRevealAPIKeyCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := resolveDeploymentID(ctx, client, c.Deployment)
	if err != nil {
		return err
	}

	resp, err := client.RevealDeploymentAPIKey(ctx, id)
	if err != nil {
		return err
	}

	out := &deploymentRevealAPIKeyOutput{APIKey: resp.APIKey}
	return c.OutputFunc(out, nil)
}

// logs

type deploymentLogsCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"logs"`

	Deployment string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Zone       v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

// show

type deploymentShowOutput struct {
	ID            v3.UUID                        `json:"id"`
	Name          string                         `json:"name"`
	Status        v3.GetDeploymentResponseStatus `json:"status"`
	GPUType       string                         `json:"gpu_type"`
	GPUCount      int64                          `json:"gpu_count"`
	Replicas      int64                          `json:"replicas"`
	ServiceLevel  string                         `json:"service_level"`
	DeploymentURL string                         `json:"deployment_url"`
	ModelID       *v3.UUID                       `json:"model_id"`
	ModelName     string                         `json:"model_name"`
	CreatedAt     string                         `json:"created_at"`
	UpdatedAt     string                         `json:"updated_at"`
}

func (o *deploymentShowOutput) ToJSON()  { output.JSON(o) }
func (o *deploymentShowOutput) ToText()  { output.Text(o) }
func (o *deploymentShowOutput) ToTable() { output.Table(o) }

type deploymentShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Deployment string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Zone       v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *deploymentShowCmd) CmdAliases() []string { return exocmd.GShowAlias }
func (c *deploymentShowCmd) CmdShort() string     { return "Show AI deployment" }
func (c *deploymentShowCmd) CmdLong() string {
	return "This command shows details of an AI deployment by ID or name."
}
func (c *deploymentShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *deploymentShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := resolveDeploymentID(ctx, client, c.Deployment)
	if err != nil {
		return err
	}

	resp, err := client.GetDeployment(ctx, id)
	if err != nil {
		return err
	}

	var modelIDPtr *v3.UUID
	var modelName string
	if resp.Model != nil {
		if resp.Model.ID.String() != "00000000-0000-0000-0000-000000000000" {
			id := resp.Model.ID
			modelIDPtr = &id
		}
		modelName = resp.Model.Name
	}

	out := &deploymentShowOutput{
		ID:            resp.ID,
		Name:          resp.Name,
		Status:        resp.Status,
		GPUType:       resp.GpuType,
		GPUCount:      resp.GpuCount,
		Replicas:      resp.Replicas,
		ServiceLevel:  resp.ServiceLevel,
		DeploymentURL: resp.DeploymentURL,
		ModelID:       modelIDPtr,
		ModelName:     modelName,
		CreatedAt:     resp.CreatedAT.Format(time.RFC3339),
		UpdatedAt:     resp.UpdatedAT.Format(time.RFC3339),
	}
	return c.OutputFunc(out, nil)
}

func (c *deploymentLogsCmd) CmdAliases() []string { return nil }
func (c *deploymentLogsCmd) CmdShort() string     { return "Get deployment logs" }
func (c *deploymentLogsCmd) CmdLong() string {
	return "This command retrieves logs for the deployment's vLLM component."
}
func (c *deploymentLogsCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *deploymentLogsCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := resolveDeploymentID(ctx, client, c.Deployment)
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
	cobra.CheckErr(exocmd.RegisterCLICommand(deploymentCmd, &deploymentListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(deploymentCmd, &deploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(deploymentCmd, &deploymentDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(deploymentCmd, &deploymentScaleCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(deploymentCmd, &deploymentRevealAPIKeyCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(deploymentCmd, &deploymentLogsCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(deploymentCmd, &deploymentShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
