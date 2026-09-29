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

type AIAPIKeyRevokeCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"revoke"`

	Keys  []string    `cli-arg:"#" cli-usage:"ID or NAME..."`
	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *AIAPIKeyRevokeCmd) CmdAliases() []string { return exocmd.GDeleteAlias }
func (c *AIAPIKeyRevokeCmd) CmdShort() string     { return "Revoke AI API key" }
func (c *AIAPIKeyRevokeCmd) CmdLong() string {
	return "This command revokes AI API keys by ID or name. Revoked keys can no longer be used."
}
func (c *AIAPIKeyRevokeCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyRevokeCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if len(c.Keys) == 0 {
		return fmt.Errorf("at least one key is required")
	}

	list, err := client.ListAIAPIKeys(ctx)
	if err != nil {
		return err
	}

	ids := []v3.UUID{}
	for _, key := range c.Keys {
		entry, err := list.FindListAIAPIKeysResponseEntry(key)
		if err != nil {
			if !c.Force {
				return err
			}
			fmt.Fprintf(os.Stderr, "warning: %s not found.\n", key)
			continue
		}

		if !c.Force {
			if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to revoke AI API key %q?", key)) {
				return nil
			}
		}

		ids = append(ids, entry.ID)
	}

	var fns []func() error
	for _, id := range ids {
		fns = append(fns, func() error {
			op, err := client.RevokeAIAPIKey(ctx, id)
			if err != nil {
				return err
			}
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			return err
		})
	}

	if err := utils.DecorateAsyncOperations("Revoking AI API key(s)...", fns...); err != nil {
		return err
	}

	if !globalstate.Quiet {
		_, _ = fmt.Fprintln(os.Stdout, "AI API key(s) revoked.")
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyRevokeCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
