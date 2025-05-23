package elastic_ip

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type elasticIPDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ElasticIP string `cli-arg:"#" cli-usage:"IP-ADDRESS|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Elastic IP zone"`
}

func (c *elasticIPDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *elasticIPDeleteCmd) CmdShort() string {
	return "Delete an Elastic IP"
}

func (c *elasticIPDeleteCmd) CmdLong() string { return "" }

func (c *elasticIPDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	elasticIP, err := globalstate.EgoscaleClient.FindElasticIP(ctx, c.Zone, c.ElasticIP)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete Elastic IP %s?", c.ElasticIP)) {
			return nil
		}
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Elastic IP %s...", c.ElasticIP), func() {
		err = globalstate.EgoscaleClient.DeleteElasticIP(ctx, c.Zone, elasticIP)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(elasticIPCmd, &elasticIPDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
