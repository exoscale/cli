package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type computeSSHKeyDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name string `cli-arg:"#"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *computeSSHKeyDeleteCmd) CmdAliases() []string { return GRemoveAlias }

func (c *computeSSHKeyDeleteCmd) CmdShort() string {
	return "Delete an SSH key"
}

func (c *computeSSHKeyDeleteCmd) CmdLong() string { return "" }

func (c *computeSSHKeyDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete SSH key %s?", c.Name)) {
			return nil
		}
	}

	err := decorateAsyncOperations(fmt.Sprintf("Deleting SSH key %s...", c.Name), func() error {
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
	cobra.CheckErr(RegisterCLICommand(computeSSHKeyCmd, &computeSSHKeyDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
