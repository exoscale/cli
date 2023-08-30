package cmd

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	ego3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/oapi"
)

type instanceResetPasswordCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reset-password"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	// TODO
	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"instance zone"`
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

	client, err := ego3.DefaultClient(ego3.ClientOptWithCredentials(
		account.CurrentAccount.Key,
		account.CurrentAccount.APISecret(),
	))
	if err != nil {
		return err
	}

	client.SetZone(oapi.ZoneName(c.Zone))

	instanceUUID, err := uuid.Parse(c.Instance)
	if err != nil {
		// TODO move this to a function named FindInstanceByName
		instanceList, err := client.Compute().Instance().List(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to list instances: %w", err)
		}

		var instanceID uuid.NullUUID
		for _, instance := range instanceList {
			if *instance.Name == c.Instance {
				instanceID = uuid.NullUUID{
					UUID:  *instance.Id,
					Valid: true,
				}
			}
		}

		if !instanceID.Valid {
			return fmt.Errorf("unable to find instance by name %q", c.Instance)
		}

		instanceUUID = instanceID.UUID
	}

	_, err = client.Compute().Instance().ResetPassword(ctx, instanceUUID)
	return err
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceResetPasswordCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
