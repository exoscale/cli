package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type iamAPIKeyDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	APIKey string `cli-arg:"#" cli-usage:"NAME|KEY"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *iamAPIKeyDeleteCmd) cmdAliases() []string { return gDeleteAlias }

func (c *iamAPIKeyDeleteCmd) cmdShort() string {
	return "Delete API Key"
}

func (c *iamAPIKeyDeleteCmd) cmdLong() string {
	return `This command deletes existing API Key.`
}

func (c *iamAPIKeyDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAPIKeyDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	if len(c.APIKey) == 27 && strings.HasPrefix(c.APIKey, "EX") {
		_, err := globalstate.EgoscaleClient.GetAPIKey(ctx, zone, c.APIKey)
		if err != nil {
			return err
		}
	} else {
		apikeys, err := globalstate.EgoscaleClient.ListAPIKeys(ctx, zone)
		if err != nil {
			return err
		}

		found := false
		for _, apikey := range apikeys {
			if apikey.Name != nil && *apikey.Name == c.APIKey {
				c.APIKey = *apikey.Key
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("key with name %q not found", c.APIKey)
		}
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete API Key %s?", c.APIKey)) {
			return nil
		}
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Deleting API Key %s...", c.APIKey), func() {
		err = globalstate.EgoscaleClient.DeleteAPIKey(ctx, zone, &egoscale.APIKey{Key: &c.APIKey})
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(iamAPIKeyCmd, &iamAPIKeyDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
