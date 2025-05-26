package ssh_key

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type computeSSHKeyDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name string `cli-arg:"#"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *computeSSHKeyDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *computeSSHKeyDeleteCmd) CmdShort() string {
	return "Delete an SSH key"
}

func (c *computeSSHKeyDeleteCmd) CmdLong() string { return "" }

func (c *computeSSHKeyDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete SSH key %s?", c.Name)) {
			return nil
		}
	}

	err := utils.DecorateAsyncOperations(fmt.Sprintf("Deleting SSH key %s...", c.Name), func() error {
		op, err := globalstate.EgoscaleV3Client.DeleteSSHKey(ctx, c.Name)
		if err != nil {
			return fmt.Errorf("exoscale: error while deleting SSH key: %w", err)
		}

		_, err = globalstate.EgoscaleV3Client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("exoscale: error while waiting for SSH key deletion: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(computeSSHKeyCmd, &computeSSHKeyDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
