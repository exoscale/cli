package apikey

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type AIAPIKeyRotateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"rotate"`

	ID string `cli-arg:"#" cli-usage:"AI API key ID"`
}

func (c *AIAPIKeyRotateCmd) CmdAliases() []string { return []string{} }
func (c *AIAPIKeyRotateCmd) CmdShort() string     { return "Rotate AI API key" }
func (c *AIAPIKeyRotateCmd) CmdLong() string {
	return "This command rotates an AI API key by ID, generating a new value."
}
func (c *AIAPIKeyRotateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyRotateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	if c.ID == "" {
		return fmt.Errorf("ID is required")
	}

	id, err := v3.ParseUUID(c.ID)
	if err != nil {
		return fmt.Errorf("invalid AI API key ID: %w", err)
	}

	resp, err := client.RotateAIAPIKey(ctx, id)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Fprintf(os.Stderr, "Store this new API key value securely.\n\n")
		fmt.Fprintf(os.Stdout, "ID:    %s\n", resp.ID)
		fmt.Fprintf(os.Stdout, "Name:  %s\n", resp.Name)
		fmt.Fprintf(os.Stdout, "Scope: %s\n", resp.Scope)
		fmt.Fprintf(os.Stdout, "Value: %s\n", resp.Value)
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyRotateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
