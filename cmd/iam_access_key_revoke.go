package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type iamAccessKeyRevokeCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"revoke"`

	APIKey string `cli-arg:"#"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *iamAccessKeyRevokeCmd) cmdAliases() []string { return gCreateAlias }

func (c *iamAccessKeyRevokeCmd) cmdShort() string {
	return "Revoke an IAM access key"
}

func (c *iamAccessKeyRevokeCmd) cmdLong() string { return "" }

func (c *iamAccessKeyRevokeCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAccessKeyRevokeCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := gCurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to revoke IAM access key IP %s?", c.APIKey)) {
			return nil
		}
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Revoking IAM access key %s...", c.APIKey), func() {
		err = globalstate.GlobalEgoscaleClient.RevokeIAMAccessKey(ctx, zone, &egoscale.IAMAccessKey{Key: &c.APIKey})
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(iamAccessKeyCmd, &iamAccessKeyRevokeCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
