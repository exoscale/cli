package dedicated_inference

import (
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

type modelListItemOutput struct {
	ID        v3.UUID                          `json:"id"`
	Name      string                           `json:"name"`
	Status    v3.ListModelsResponseEntryStatus `json:"status"`
	ModelSize *int64                           `json:"model_size"`
	CreatedAt string                           `json:"created_at"`
	UpdatedAt string                           `json:"updated_at"`
}

type modelListOutput []modelListItemOutput

func (o *modelListOutput) ToJSON()  { output.JSON(o) }
func (o *modelListOutput) ToText()  { output.Text(o) }
func (o *modelListOutput) ToTable() { output.Table(o) }

type modelListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *modelListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *modelListCmd) CmdShort() string     { return "List AI models" }
func (c *modelListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists AI models.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&modelListOutput{}), ", "))
}
func (c *modelListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *modelListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	resp, err := client.ListModels(ctx)
	if err != nil {
		return err
	}

	out := make(modelListOutput, 0, len(resp.Models))
	for _, m := range resp.Models {
		var sizePtr *int64
		if m.ModelSize != 0 {
			v := m.ModelSize
			sizePtr = &v
		}
		out = append(out, modelListItemOutput{
			ID:        m.ID,
			Name:      m.Name,
			Status:    m.Status,
			ModelSize: sizePtr,
			CreatedAt: m.CreatedAT.Format(time.RFC3339),
			UpdatedAt: m.UpdatedAT.Format(time.RFC3339),
		})
	}

	return c.OutputFunc(&out, nil)
}

// create

type modelCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name             string `cli-flag:"name" cli-usage:"Model name (e.g. openai/gpt-oss-120b)"`
	HuggingfaceToken string `cli-flag:"huggingface-token" cli-usage:"Huggingface token if required by the model"`
}

func (c *modelCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }
func (c *modelCreateCmd) CmdShort() string { return "Create AI model (download from Huggingface)" }
func (c *modelCreateCmd) CmdLong() string {
	return "This command creates an AI model by downloading it from Huggingface."
}
func (c *modelCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *modelCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	if c.Name == "" {
		return fmt.Errorf("--name is required")
	}

	req := v3.CreateModelRequest{
		Name:             c.Name,
		HuggingfaceToken: c.HuggingfaceToken,
	}

	var op *v3.Operation
	var err error
	utils.DecorateAsyncOperation(fmt.Sprintf("Creating model %q...", c.Name), func() {
		op, err = client.CreateModel(ctx, req)
		if err != nil {
			return
		}
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}
	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Model creation initiated.")
	}
	return nil
}

// delete

type modelDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ID string `cli-arg:"#" cli-usage:"MODEL-ID (UUID)"`
}

func (c *modelDeleteCmd) CmdAliases() []string { return exocmd.GDeleteAlias }
func (c *modelDeleteCmd) CmdShort() string     { return "Delete AI model" }
func (c *modelDeleteCmd) CmdLong() string      { return "This command deletes an AI model by its ID." }
func (c *modelDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *modelDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	id, err := v3.ParseUUID(c.ID)
	if err != nil {
		return fmt.Errorf("invalid model ID: %w", err)
	}

	var op *v3.Operation
	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting model %s...", c.ID), func() {
		op, err = client.DeleteModel(ctx, id)
		if err != nil {
			return
		}
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}
	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Model deletion initiated.")
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(modelCmd, &modelListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(modelCmd, &modelCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
	cobra.CheckErr(exocmd.RegisterCLICommand(modelCmd, &modelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
