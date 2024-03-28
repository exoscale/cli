package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceResetPasswordCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reset-password"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceResetPasswordCmd) cmdAliases() []string { return nil }

func (c *instanceResetPasswordCmd) cmdShort() string {
	return "Reset the password of a Compute instance"
}

func (c *instanceResetPasswordCmd) cmdLong() string { return "" }

func (c *instanceResetPasswordCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceResetPasswordCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	_, err = globalstate.EgoscaleClient.ResetInstancePasswordWithResponse(ctx, *instance.ID)
	return err
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceResetPasswordCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
