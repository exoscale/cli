package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type computeSSHKeyDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name string `cli-arg:"#"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *computeSSHKeyDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *computeSSHKeyDeleteCmd) cmdShort() string {
	return "Delete an SSH key"
}

func (c *computeSSHKeyDeleteCmd) cmdLong() string { return "" }

func (c *computeSSHKeyDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {

	ctx := gContext

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete SSH key %s?", c.Name)) {
			return nil
		}
	}

	decorateAsyncOperations(fmt.Sprintf("Deleting SSH key %s...", c.Name), func() error {
		op, err := globalstate.EgoscaleV3Client.DeleteSSHKey(ctx, c.Name)

		if err != nil {
			return err
		}

		 _, err = globalstate.EgoscaleV3Client.Wait(ctx, op, v3.OperationStateSuccess)
		 if err != nil {
			return err
		}

		return nil
	})

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeSSHKeyCmd, &computeSSHKeyDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
