package apikey

import (
	"fmt"
	"os"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type AIAPIKeyUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	ID    string `cli-arg:"#" cli-usage:"AI API key ID"`
	Name  string `cli-flag:"name" cli-usage:"AI API key name"`
	Scope string `cli-flag:"scope" cli-usage:"Scope: 'public' for all deployments, or a deployment UUID"`
}

func (c *AIAPIKeyUpdateCmd) CmdAliases() []string { return exocmd.GUpdateAlias }
func (c *AIAPIKeyUpdateCmd) CmdShort() string     { return "Update AI API key" }
func (c *AIAPIKeyUpdateCmd) CmdLong() string {
	return "This command updates an AI API key by ID."
}
func (c *AIAPIKeyUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyUpdateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	if c.ID == "" {
		return fmt.Errorf("ID is required")
	}
	if c.Name == "" && c.Scope == "" {
		return fmt.Errorf("at least one of --name or --scope is required")
	}

	id, err := v3.ParseUUID(c.ID)
	if err != nil {
		return fmt.Errorf("invalid AI API key ID: %w", err)
	}

	req := v3.UpdateAIAPIKeyRequest{}
	if c.Name != "" {
		req.Name = c.Name
	}
	if c.Scope != "" {
		req.Scope = c.Scope
	}

	resp, err := client.UpdateAIAPIKey(ctx, id, req)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Fprintf(os.Stdout, "AI API key updated.\n")
		fmt.Fprintf(os.Stdout, "ID:        %s\n", resp.ID)
		fmt.Fprintf(os.Stdout, "Name:      %s\n", resp.Name)
		fmt.Fprintf(os.Stdout, "Scope:     %s\n", resp.Scope)
		fmt.Fprintf(os.Stdout, "Updated at: %s\n", resp.UpdatedAT.Format(time.RFC3339))
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyUpdateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
