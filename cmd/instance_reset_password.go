package cmd

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/client"
	"github.com/exoscale/cli/pkg/instance"
	"github.com/exoscale/egoscale/v3/oapi"
)

type instanceResetPasswordCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reset-password"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceResetPasswordCmd) cmdAliases() []string { return gRemoveAlias }

func (c *instanceResetPasswordCmd) cmdShort() string {
	return "Reset the password of a Compute instance"
}

func (c *instanceResetPasswordCmd) cmdLong() string { return "" }

func (c *instanceResetPasswordCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceResetPasswordCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var (
		err error
		ctx = context.Background()
	)

	v3Client, err := client.Get()
	if err != nil {
		return err
	}

	v3Client.SetZone(oapi.ZoneName(c.Zone))

	instanceUUID, err := uuid.Parse(c.Instance)
	if err != nil {
		instance, err := instance.FindInstanceByName(ctx, &v3Client.Client, c.Instance)
		if err != nil {
			return err
		}

		if instance == nil {
			return fmt.Errorf("unable to find instance by name %q", c.Instance)
		}

		instanceUUID = *instance.Id
	}

	_, err = v3Client.Compute().Instance().ResetPassword(ctx, instanceUUID)
	return err
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceResetPasswordCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
