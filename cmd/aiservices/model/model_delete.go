package model

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type ModelDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	IDs   []string    `cli-arg:"#" cli-usage:"MODEL-ID (UUID)..."`
	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *ModelDeleteCmd) CmdAliases() []string { return exocmd.GDeleteAlias }
func (c *ModelDeleteCmd) CmdShort() string     { return "Delete AI model" }
func (c *ModelDeleteCmd) CmdLong() string      { return "This command deletes an AI model by its ID." }
func (c *ModelDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *ModelDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	modelsToDelete := []v3.UUID{}
	for _, idStr := range c.IDs {
		id, err := v3.ParseUUID(idStr)
		if err != nil {
			if !c.Force {
				return fmt.Errorf("invalid model ID %q: %w", idStr, err)
			}
			fmt.Fprintf(os.Stderr, "warning: invalid model ID %q: %v\n", idStr, err)
			continue
		}

		if !c.Force {
			if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete model %q?", idStr)) {
				return nil
			}
		}

		modelsToDelete = append(modelsToDelete, id)
	}

	var fns []func() error
	for _, id := range modelsToDelete {
		fns = append(fns, func() error {
			op, err := client.DeleteModel(ctx, id)
			if err != nil {
				return err
			}
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			return err
		})
	}

	err = utils.DecorateAsyncOperations(fmt.Sprintf("Deleting model(s)..."), fns...)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Model(s) deleted.")
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
