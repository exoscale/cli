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

type ModelDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ID   string      `cli-arg:"#" cli-usage:"MODEL-ID (UUID)"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
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
    cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
