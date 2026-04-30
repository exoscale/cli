package apikey

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type AIAPIKeyCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name  string `cli-flag:"name" cli-usage:"AI API key name"`
	Scope string `cli-flag:"scope" cli-usage:"Scope: 'public' for all deployments, or a deployment UUID"`
}

func (c *AIAPIKeyCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }
func (c *AIAPIKeyCreateCmd) CmdShort() string     { return "Create AI API key" }
func (c *AIAPIKeyCreateCmd) CmdLong() string {
	return "This command creates an AI API key."
}
func (c *AIAPIKeyCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	if c.Name == "" {
		return fmt.Errorf("--name is required")
	}
	if c.Scope == "" {
		return fmt.Errorf("--scope is required")
	}

	req := v3.CreateAIAPIKeyRequest{
		Name:  c.Name,
		Scope: c.Scope,
	}

	resp, err := client.CreateAIAPIKey(ctx, req)
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
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
