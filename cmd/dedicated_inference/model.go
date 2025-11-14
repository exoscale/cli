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
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage AI models",
}

func init() {
	DedicatedInferenceCmd.AddCommand(modelCmd)
}

// int64PtrIfNonZero returns a pointer to v if it's non-zero, otherwise nil.
func int64PtrIfNonZero(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

type modelListItemOutput struct {
	ID        v3.UUID                          `json:"id"`
	Name      string                           `json:"name"`
	Status    v3.ListModelsResponseEntryStatus `json:"status"`
	ModelSize *int64                           `json:"model_size"`
}

type modelListOutput []modelListItemOutput

func (o *modelListOutput) ToJSON()  { output.JSON(o) }
func (o *modelListOutput) ToText()  { output.Text(o) }
func (o *modelListOutput) ToTable() { output.Table(o) }

// show

type modelShowOutput struct {
	ID        v3.UUID                   `json:"id"`
	Name      string                    `json:"name"`
	Status    v3.GetModelResponseStatus `json:"status"`
	ModelSize *int64                    `json:"model_size"`
	CreatedAt string                    `json:"created_at"`
	UpdatedAt string                    `json:"updated_at"`
}

func (o *modelShowOutput) ToJSON()  { output.JSON(o) }
func (o *modelShowOutput) ToText()  { output.Text(o) }
func (o *modelShowOutput) ToTable() { output.Table(o) }

type modelShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	ID   string      `cli-arg:"#" cli-usage:"MODEL-ID (UUID)"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *modelShowCmd) CmdAliases() []string { return exocmd.GShowAlias }
func (c *modelShowCmd) CmdShort() string     { return "Show AI model" }
func (c *modelShowCmd) CmdLong() string {
	return "This command shows details of an AI model by its ID."
}
func (c *modelShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *modelShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := v3.ParseUUID(c.ID)
	if err != nil {
		return fmt.Errorf("invalid model ID: %w", err)
	}
	resp, err := client.GetModel(ctx, id)
	if err != nil {
		return err
	}
	sizePtr := int64PtrIfNonZero(resp.ModelSize)
	out := &modelShowOutput{
		ID:        resp.ID,
		Name:      resp.Name,
		Status:    resp.Status,
		ModelSize: sizePtr,
		CreatedAt: resp.CreatedAT.Format(time.RFC3339),
		UpdatedAt: resp.UpdatedAT.Format(time.RFC3339),
	}
	return c.OutputFunc(out, nil)
}

type modelListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *modelListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *modelListCmd) CmdShort() string     { return "List AI models" }
func (c *modelListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists AI models.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&modelListOutput{}), ", "))
}
func (c *modelListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *modelListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListModels(ctx)
	if err != nil {
		return err
	}

	out := make(modelListOutput, 0, len(resp.Models))
	for _, m := range resp.Models {
		sizePtr := int64PtrIfNonZero(m.ModelSize)
		out = append(out, modelListItemOutput{
			ID:        m.ID,
			Name:      m.Name,
			Status:    m.Status,
			ModelSize: sizePtr,
		})
	}

	return c.OutputFunc(&out, nil)
}

// create

type modelCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name             string      `cli-flag:"name" cli-usage:"Model name (e.g. openai/gpt-oss-120b)"`
	HuggingfaceToken string      `cli-flag:"huggingface-token" cli-usage:"Huggingface token if required by the model"`
	Zone             v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *modelCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }
func (c *modelCreateCmd) CmdShort() string     { return "Create AI model (download from Huggingface)" }
func (c *modelCreateCmd) CmdLong() string {
	return "This command creates an AI model by downloading it from Huggingface."
}
func (c *modelCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *modelCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if c.Name == "" {
		return fmt.Errorf("--name is required")
	}

	req := v3.CreateModelRequest{
		Name:             c.Name,
		HuggingfaceToken: c.HuggingfaceToken,
	}

	if err := utils.RunAsync(ctx, client, fmt.Sprintf("Creating model %q...", c.Name), func(ctx context.Context, c *v3.Client) (*v3.Operation, error) {
		return c.CreateModel(ctx, req)
	}); err != nil {
		return err
	}
	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Model created.")
	}
	return nil
}

// delete

type modelDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ID   string      `cli-arg:"#" cli-usage:"MODEL-ID (UUID)"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *modelDeleteCmd) CmdAliases() []string { return exocmd.GDeleteAlias }
func (c *modelDeleteCmd) CmdShort() string     { return "Delete AI model" }
func (c *modelDeleteCmd) CmdLong() string      { return "This command deletes an AI model by its ID." }
func (c *modelDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *modelDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := v3.ParseUUID(c.ID)
	if err != nil {
		return fmt.Errorf("invalid model ID: %w", err)
	}

	if err := utils.RunAsync(ctx, client, fmt.Sprintf("Deleting model %s...", c.ID), func(ctx context.Context, c *v3.Client) (*v3.Operation, error) {
		return c.DeleteModel(ctx, id)
	}); err != nil {
		return err
	}
	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Model deleted.")
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(modelCmd, &modelListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(modelCmd, &modelCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(modelCmd, &modelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(modelCmd, &modelShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
