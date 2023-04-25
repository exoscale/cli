package cmd

import (
	"errors"
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type elasticIPDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ElasticIP string `cli-arg:"#" cli-usage:"IP-ADDRESS|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Elastic IP zone"`
}

func (c *elasticIPDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *elasticIPDeleteCmd) cmdShort() string {
	return "Delete an Elastic IP"
}

func (c *elasticIPDeleteCmd) cmdLong() string { return "" }

func (c *elasticIPDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	elasticIP, err := globalstate.GlobalEgoscaleClient.FindElasticIP(ctx, c.Zone, c.ElasticIP)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Elastic IP %s?", c.ElasticIP)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Elastic IP %s...", c.ElasticIP), func() {
		err = globalstate.GlobalEgoscaleClient.DeleteElasticIP(ctx, c.Zone, elasticIP)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(elasticIPCmd, &elasticIPDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
