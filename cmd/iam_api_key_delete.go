package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type iamAPIKeyDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	APIKey string `cli-arg:"#" cli-usage:"NAME|KEY"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *iamAPIKeyDeleteCmd) CmdAliases() []string { return GDeleteAlias }

func (c *iamAPIKeyDeleteCmd) CmdShort() string {
	return "Delete an API Key"
}

func (c *iamAPIKeyDeleteCmd) CmdLong() string {
	return `This command deletes existing API Key.`
}

func (c *iamAPIKeyDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAPIKeyDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client := globalstate.EgoscaleV3Client

	listAPIKeysResp, err := client.ListAPIKeys(ctx)
	if err != nil {
		return err
	}

	apiKey, err := listAPIKeysResp.FindIAMAPIKey(c.APIKey)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete API Key %q?", c.APIKey)) {
			return nil
		}
	}

	return decorateAsyncOperations(fmt.Sprintf("Deleting API Key %s...", c.APIKey), func() error {
		op, err := client.DeleteAPIKey(ctx, apiKey.Key)
		if err != nil {
			return fmt.Errorf("exoscale: error while deleting IAM API Key: %w", err)
		}

		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("exoscale: error while waiting for IAM API Key deletion: %w", err)
		}

		return nil
	})
}

func init() {
	cobra.CheckErr(RegisterCLICommand(iamAPIKeyCmd, &iamAPIKeyDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
