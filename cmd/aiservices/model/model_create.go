package model

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

type ModelCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name             string      `cli-arg:"#" cli-usage:"NAME (e.g. swiss-ai/Apertus-8B-Instruct-2509)"`
	HuggingfaceToken string      `cli-flag:"huggingface-token" cli-usage:"Huggingface token if required by the model"`
	Zone             v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *ModelCreateCmd) CmdAliases() []string { return append(exocmd.GCreateAlias, "download") }
func (c *ModelCreateCmd) CmdShort() string     { return "Create AI model (download from Huggingface)" }
func (c *ModelCreateCmd) CmdLong() string {
	return "This command creates an AI model by downloading it from Huggingface.\n\n" +
		"The name parameter must be a valid Huggingface model name (e.g. mistralai/Mixtral-8x7B-Instruct-v0.1)."
}
func (c *ModelCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *ModelCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if c.Name == "" {
		return fmt.Errorf("NAME is required")
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

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &ModelCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
