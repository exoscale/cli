package apikey

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type AIAPIKeyRevealCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reveal"`

	ID string `cli-arg:"#" cli-usage:"AI API key ID"`
}

func (c *AIAPIKeyRevealCmd) CmdAliases() []string { return []string{} }
func (c *AIAPIKeyRevealCmd) CmdShort() string     { return "Reveal AI API key value" }
func (c *AIAPIKeyRevealCmd) CmdLong() string {
	return "This command reveals the secret value of an AI API key by ID."
}
func (c *AIAPIKeyRevealCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyRevealCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	if c.ID == "" {
		return fmt.Errorf("ID is required")
	}

	id, err := v3.ParseUUID(c.ID)
	if err != nil {
		return fmt.Errorf("invalid AI API key ID: %w", err)
	}

	resp, err := client.RevealAIAPIKey(ctx, id)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Fprintf(os.Stderr, "Store this API key value securely.\n\n")
		fmt.Fprintf(os.Stdout, "ID:    %s\n", resp.ID)
		fmt.Fprintf(os.Stdout, "Name:  %s\n", resp.Name)
		fmt.Fprintf(os.Stdout, "Scope: %s\n", resp.Scope)
		fmt.Fprintf(os.Stdout, "Value: %s\n", resp.Value)
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyRevealCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
