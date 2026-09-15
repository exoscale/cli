package apikey

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type AIAPIKeyDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	IDs   []string `cli-arg:"#" cli-usage:"AI API key ID..."`
	Force bool     `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *AIAPIKeyDeleteCmd) CmdAliases() []string { return exocmd.GDeleteAlias }
func (c *AIAPIKeyDeleteCmd) CmdShort() string     { return "Delete AI API key" }
func (c *AIAPIKeyDeleteCmd) CmdLong() string {
	return "This command deletes AI API keys by ID."
}
func (c *AIAPIKeyDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	if len(c.IDs) == 0 {
		return fmt.Errorf("at least one ID is required")
	}

	for _, idStr := range c.IDs {
		id, err := v3.ParseUUID(idStr)
		if err != nil {
			if !c.Force {
				return fmt.Errorf("invalid AI API key ID %q: %w", idStr, err)
			}
			fmt.Fprintf(os.Stderr, "warning: invalid AI API key ID %q\n", idStr)
			continue
		}

		if !c.Force {
			if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete AI API key %q?", idStr)) {
				continue
			}
		}

		if _, err := client.DeleteAIAPIKey(ctx, id); err != nil {
			if !c.Force {
				return err
			}
			fmt.Fprintf(os.Stderr, "warning: failed to delete AI API key %q: %v\n", idStr, err)
			continue
		}
	}

	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "AI API key(s) deleted.")
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
